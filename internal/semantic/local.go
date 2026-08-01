package semantic

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"context"
	"math"
	"os"
	"strings"
)

// LocalClient computes word similarity using pre-computed embeddings loaded
// from disk. No network calls, no Python dependency at runtime.
//
// The embeddings are generated once by scripts/precompute.py and stored as:
//   - vocabulary.json: JSON array of words ([]string)
//   - embeddings.npy:  NumPy .npy file, float32, shape [N, embeddingDim]
type LocalClient struct {
	vocab      []string
	embeddings [][]float32
	index      map[string]int
	dim        int
}

// NewLocal loads pre-computed embeddings from the given paths and returns
// a ready-to-use LocalClient that implements the Semantic interface.
func NewLocal(vocabPath, embeddingsPath string) (*LocalClient, error) {
	// ── 1. Load vocabulary ──
	vocabData, err := os.ReadFile(vocabPath)
	if err != nil {
		return nil, fmt.Errorf("read vocabulary: %w", err)
	}

	var vocab []string
	if err := json.Unmarshal(vocabData, &vocab); err != nil {
		return nil, fmt.Errorf("parse vocabulary: %w", err)
	}

	if len(vocab) == 0 {
		return nil, fmt.Errorf("vocabulary is empty")
	}

	// ── 2. Load embeddings (.npy format) ──
	embeddings, dim, err := readNpy(embeddingsPath)
	if err != nil {
		return nil, fmt.Errorf("read embeddings: %w", err)
	}

	if len(embeddings) != len(vocab) {
		return nil, fmt.Errorf("vocab/embeddings size mismatch: %d words vs %d vectors", len(vocab), len(embeddings))
	}

	// ── 3. Build lookup index ──
	idx := make(map[string]int, len(vocab))
	for i, w := range vocab {
		idx[strings.ToLower(w)] = i
	}

	return &LocalClient{
		vocab:      vocab,
		embeddings: embeddings,
		index:      idx,
		dim:        dim,
	}, nil
}

// Similarity implements the Semantic interface.
// Returns cosine similarity between the two words' embeddings.
// If either word is unknown (not in the vocabulary), returns 0.0 with no error.
func (lc *LocalClient) Similarity(_ context.Context, word, target string) (float64, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	target = strings.ToLower(strings.TrimSpace(target))

	i, ok1 := lc.index[word]
	j, ok2 := lc.index[target]

	if !ok1 || !ok2 {
		return 0.0, nil
	}

	return cosineSimilarity(lc.embeddings[i], lc.embeddings[j]), nil
}

// VocabSize returns the number of words in the vocabulary.
func (lc *LocalClient) VocabSize() int {
	return len(lc.vocab)
}

// ──────────────────────────────────────────────
// NumPy .npy reader (minimal, float32 only)
// ──────────────────────────────────────────────
//
// .npy format (v1.0 / v2.0):
//   Bytes 0-5:   magic "\x93NUMPY"
//   Byte  6:     major version
//   Byte  7:     minor version
//   Bytes 8-9:   header length (uint16 LE for v1, uint32 LE for v2)
//   header:      ASCII dict with 'descr', 'fortran_order', 'shape'
//   data:        raw binary, row-major

func readNpy(path string) ([][]float32, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	// Validate magic
	magic := []byte{0x93, 'N', 'U', 'M', 'P', 'Y'}
	if len(data) < 10 || string(data[:6]) != string(magic) {
		return nil, 0, fmt.Errorf("not a valid .npy file")
	}

	majorVersion := data[6]

	var headerLen int
	var headerStart int

	switch majorVersion {
	case 1:
		headerLen = int(binary.LittleEndian.Uint16(data[8:10]))
		headerStart = 10
	case 2:
		headerLen = int(binary.LittleEndian.Uint32(data[8:12]))
		headerStart = 12
	default:
		return nil, 0, fmt.Errorf("unsupported npy version: %d", majorVersion)
	}

	header := string(data[headerStart : headerStart+headerLen])
	dataStart := headerStart + headerLen

	// Parse shape from header dict — e.g. "'shape': (3000, 384)"
	rows, cols, err := parseShape(header)
	if err != nil {
		return nil, 0, err
	}

	// Verify we have enough data
	expectedBytes := rows * cols * 4 // float32 = 4 bytes
	if len(data)-dataStart < expectedBytes {
		return nil, 0, fmt.Errorf("npy file truncated: expected %d bytes of data, got %d",
			expectedBytes, len(data)-dataStart)
	}

	// Read raw float32 values (little-endian)
	embeddings := make([][]float32, rows)
	offset := dataStart
	for i := 0; i < rows; i++ {
		row := make([]float32, cols)
		for j := 0; j < cols; j++ {
			bits := binary.LittleEndian.Uint32(data[offset : offset+4])
			row[j] = math.Float32frombits(bits)
			offset += 4
		}
		embeddings[i] = row
	}

	return embeddings, cols, nil
}

// parseShape extracts (rows, cols) from the npy header string.
// Header looks like: "{'descr': '<f4', 'fortran_order': False, 'shape': (3000, 384), }\n"
func parseShape(header string) (int, int, error) {
	// Find "shape"
	idx := strings.Index(header, "'shape'")
	if idx < 0 {
		idx = strings.Index(header, "\"shape\"")
	}
	if idx < 0 {
		return 0, 0, fmt.Errorf("shape not found in npy header")
	}

	// Find the opening parenthesis after "shape"
	rest := header[idx:]
	openParen := strings.Index(rest, "(")
	closeParen := strings.Index(rest, ")")
	if openParen < 0 || closeParen < 0 || closeParen <= openParen {
		return 0, 0, fmt.Errorf("malformed shape in npy header")
	}

	shapePart := rest[openParen+1 : closeParen]
	shapePart = strings.TrimSpace(shapePart)

	var rows, cols int
	n, err := fmt.Sscanf(shapePart, "%d, %d", &rows, &cols)
	if err != nil || n != 2 {
		return 0, 0, fmt.Errorf("failed to parse shape (%s): %v", shapePart, err)
	}

	return rows, cols, nil
}

// ──────────────────────────────────────────────
// Cosine similarity
// ──────────────────────────────────────────────

func cosineSimilarity(a, b []float32) float64 {
	var dot, normA, normB float64
	for i := range a {
		ai := float64(a[i])
		bi := float64(b[i])
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
