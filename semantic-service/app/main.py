from fastapi import FastAPI 
from pydantic import BaseModel

from .model import embed
from .similarity import cosine_similarity


app = FastAPI()

class SimilarityRequest(BaseModel):
    word:   str
    target: str
    
    
    
@app.post("/similarity")
def similarity(req: SimilarityRequest):
    v1 = embed(req.word)
    v2 = embed(req.target)
    score = cosine_similarity(v1, v2)
    return {"similarity": score}



