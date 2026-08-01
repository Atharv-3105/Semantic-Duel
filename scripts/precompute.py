import json
import os
import time

import numpy as np
from sentence_transformers import SentenceTransformer

# ──────────────────────────────────────────────
# Curated vocabulary: 3000+ common English words
# Categories: nature, emotions, actions, objects, places, science,
#             food, body, relationships, time, movement, abstract,
#             weather, colors, animals, professions, sports, music,
#             technology, materials, clothing, transportation, etc.
# ──────────────────────────────────────────────

VOCABULARY = sorted(set([
    # ─── Nature & Environment ───
    "ocean", "river", "lake", "stream", "pond", "waterfall", "creek", "spring",
    "mountain", "hill", "valley", "canyon", "cliff", "ridge", "peak", "summit",
    "forest", "jungle", "woodland", "grove", "meadow", "field", "prairie", "plain",
    "desert", "dune", "oasis", "savanna", "tundra", "glacier", "iceberg", "frost",
    "island", "peninsula", "continent", "shore", "beach", "coast", "reef", "coral",
    "volcano", "lava", "magma", "crater", "earthquake", "tremor", "avalanche",
    "cave", "cavern", "grotto", "rock", "stone", "boulder", "pebble", "gravel",
    "sand", "mud", "clay", "soil", "dirt", "earth", "ground", "terrain",
    "sky", "cloud", "fog", "mist", "haze", "rainbow", "aurora", "horizon",
    "sun", "moon", "star", "planet", "galaxy", "universe", "cosmos", "nebula",
    "tree", "leaf", "branch", "root", "trunk", "bark", "seed", "sprout",
    "flower", "petal", "blossom", "bloom", "bud", "thorn", "vine", "moss",
    "grass", "weed", "fern", "bush", "shrub", "hedge", "garden", "orchard",
    "marsh", "swamp", "bog", "wetland", "delta", "estuary", "lagoon", "bay",
    "tide", "wave", "current", "rapids", "whirlpool", "geyser", "hot spring",
    "mineral", "crystal", "gem", "diamond", "ruby", "emerald", "sapphire", "quartz",

    # ─── Weather & Seasons ───
    "rain", "snow", "sleet", "hail", "drizzle", "downpour", "shower", "storm",
    "thunder", "lightning", "tornado", "hurricane", "cyclone", "typhoon", "blizzard",
    "wind", "breeze", "gust", "gale", "draft", "zephyr", "monsoon", "squall",
    "sunshine", "shade", "shadow", "dusk", "dawn", "twilight", "sunset", "sunrise",
    "winter", "spring", "summer", "autumn", "fall", "season", "solstice", "equinox",
    "temperature", "humidity", "pressure", "climate", "weather", "forecast", "drought",
    "flood", "tsunami", "mudslide", "wildfire", "heatwave", "coldwave", "icicle",

    # ─── Animals ───
    "dog", "cat", "horse", "cow", "pig", "sheep", "goat", "chicken",
    "duck", "goose", "turkey", "rabbit", "mouse", "rat", "hamster", "guinea",
    "lion", "tiger", "bear", "wolf", "fox", "deer", "elk", "moose",
    "elephant", "giraffe", "zebra", "hippo", "rhino", "gorilla", "monkey", "chimpanzee",
    "eagle", "hawk", "falcon", "owl", "crow", "raven", "sparrow", "robin",
    "parrot", "penguin", "flamingo", "swan", "heron", "pelican", "seagull", "dove",
    "snake", "lizard", "turtle", "crocodile", "alligator", "frog", "toad", "salamander",
    "fish", "shark", "whale", "dolphin", "octopus", "squid", "jellyfish", "starfish",
    "crab", "lobster", "shrimp", "clam", "oyster", "snail", "slug", "worm",
    "butterfly", "moth", "bee", "wasp", "ant", "beetle", "spider", "scorpion",
    "dragonfly", "firefly", "cricket", "grasshopper", "ladybug", "caterpillar", "cocoon",
    "panther", "leopard", "cheetah", "jaguar", "cougar", "lynx", "bobcat", "hyena",
    "buffalo", "bison", "camel", "llama", "alpaca", "donkey", "mule", "stallion",
    "puppy", "kitten", "cub", "foal", "calf", "lamb", "piglet", "chick",
    "salmon", "trout", "bass", "tuna", "cod", "swordfish", "ray", "eel",
    "seal", "walrus", "otter", "beaver", "badger", "raccoon", "skunk", "porcupine",

    # ─── Emotions & Feelings ───
    "happy", "sad", "angry", "afraid", "scared", "anxious", "nervous", "worried",
    "excited", "thrilled", "elated", "joyful", "cheerful", "merry", "blissful", "ecstatic",
    "calm", "peaceful", "serene", "tranquil", "relaxed", "content", "satisfied", "pleased",
    "love", "affection", "fondness", "adoration", "devotion", "passion", "desire", "longing",
    "hate", "disgust", "contempt", "loathing", "resentment", "bitterness", "spite", "malice",
    "fear", "terror", "horror", "dread", "panic", "fright", "alarm", "shock",
    "surprise", "amazement", "wonder", "awe", "astonishment", "bewilderment", "confusion",
    "grief", "sorrow", "mourning", "heartbreak", "despair", "agony", "anguish", "misery",
    "pride", "confidence", "dignity", "honor", "shame", "guilt", "regret", "remorse",
    "hope", "optimism", "faith", "trust", "doubt", "skepticism", "cynicism", "pessimism",
    "jealousy", "envy", "greed", "lust", "gluttony", "sloth", "wrath", "vanity",
    "courage", "bravery", "valor", "heroism", "cowardice", "timidity", "humility", "modesty",
    "sympathy", "empathy", "compassion", "pity", "mercy", "kindness", "generosity", "gratitude",
    "loneliness", "isolation", "solitude", "abandonment", "rejection", "betrayal", "nostalgia",
    "frustration", "irritation", "annoyance", "impatience", "rage", "fury", "indignation",
    "enthusiasm", "eagerness", "zeal", "fervor", "inspiration", "motivation", "determination",

    # ─── Actions & Verbs ───
    "run", "walk", "jump", "climb", "crawl", "swim", "fly", "dive",
    "sit", "stand", "lie", "kneel", "crouch", "lean", "bend", "stretch",
    "push", "pull", "lift", "carry", "drag", "drop", "throw", "catch",
    "hit", "kick", "punch", "slap", "strike", "knock", "pound", "smash",
    "cut", "slice", "chop", "carve", "trim", "shave", "peel", "scrape",
    "eat", "drink", "swallow", "chew", "bite", "sip", "gulp", "taste",
    "cook", "bake", "fry", "grill", "roast", "boil", "steam", "simmer",
    "read", "write", "draw", "paint", "sketch", "type", "print", "erase",
    "speak", "talk", "whisper", "shout", "yell", "scream", "sing", "hum",
    "listen", "hear", "watch", "look", "see", "stare", "glance", "gaze",
    "think", "remember", "forget", "learn", "teach", "study", "practice", "train",
    "build", "create", "design", "invent", "construct", "assemble", "install", "repair",
    "break", "destroy", "demolish", "crush", "shatter", "tear", "rip", "crack",
    "open", "close", "lock", "unlock", "seal", "wrap", "unwrap", "fold",
    "give", "take", "share", "lend", "borrow", "steal", "grab", "snatch",
    "buy", "sell", "trade", "exchange", "bargain", "negotiate", "invest", "spend",
    "search", "find", "discover", "explore", "investigate", "examine", "inspect", "analyze",
    "hide", "reveal", "show", "display", "present", "demonstrate", "exhibit", "expose",
    "start", "begin", "launch", "initiate", "activate", "trigger", "ignite", "spark",
    "stop", "end", "finish", "complete", "conclude", "terminate", "halt", "cease",
    "wait", "pause", "delay", "hesitate", "linger", "loiter", "stall", "procrastinate",
    "move", "travel", "wander", "roam", "drift", "float", "glide", "soar",
    "dance", "spin", "twirl", "rotate", "revolve", "orbit", "circle", "spiral",
    "grow", "shrink", "expand", "contract", "swell", "inflate", "deflate", "dissolve",
    "mix", "blend", "combine", "merge", "separate", "divide", "split", "scatter",
    "pour", "spill", "drip", "leak", "spray", "squirt", "splash", "soak",
    "burn", "melt", "freeze", "evaporate", "condense", "solidify", "crystallize", "decay",
    "breathe", "inhale", "exhale", "cough", "sneeze", "yawn", "sigh", "gasp",
    "sleep", "wake", "dream", "rest", "nap", "doze", "snore", "meditate",
    "laugh", "cry", "smile", "frown", "wink", "blink", "nod", "shake",
    "hug", "kiss", "embrace", "cuddle", "caress", "touch", "feel", "sense",
    "heal", "cure", "treat", "recover", "survive", "thrive", "flourish", "prosper",
    "fight", "battle", "struggle", "compete", "challenge", "confront", "oppose", "resist",
    "help", "assist", "support", "guide", "lead", "follow", "obey", "command",
    "protect", "defend", "guard", "shield", "shelter", "rescue", "save", "preserve",
    "attack", "invade", "raid", "ambush", "siege", "conquer", "capture", "surrender",

    # ─── Objects & Things ───
    "book", "pen", "pencil", "paper", "notebook", "journal", "diary", "letter",
    "phone", "computer", "laptop", "tablet", "keyboard", "mouse", "screen", "monitor",
    "chair", "table", "desk", "bench", "stool", "couch", "sofa", "bed",
    "door", "window", "wall", "floor", "ceiling", "roof", "stair", "ramp",
    "lamp", "candle", "lantern", "torch", "flashlight", "bulb", "chandelier", "spotlight",
    "mirror", "glass", "cup", "mug", "bottle", "jar", "bowl", "plate",
    "knife", "fork", "spoon", "chopstick", "whisk", "ladle", "spatula", "tongs",
    "pot", "pan", "kettle", "oven", "stove", "microwave", "toaster", "blender",
    "clock", "watch", "alarm", "timer", "calendar", "compass", "telescope", "microscope",
    "camera", "lens", "film", "photograph", "picture", "painting", "portrait", "sculpture",
    "key", "lock", "chain", "rope", "wire", "cable", "cord", "string",
    "bag", "box", "basket", "crate", "barrel", "trunk", "chest", "case",
    "coin", "bill", "wallet", "purse", "safe", "vault", "treasure", "jewel",
    "ring", "necklace", "bracelet", "earring", "brooch", "crown", "tiara", "medal",
    "flag", "banner", "sign", "poster", "billboard", "label", "tag", "stamp",
    "tool", "hammer", "wrench", "screwdriver", "drill", "saw", "pliers", "clamp",
    "nail", "screw", "bolt", "nut", "pin", "needle", "thread", "ribbon",
    "brush", "comb", "razor", "scissors", "tweezers", "file", "sandpaper", "polish",
    "umbrella", "fan", "heater", "radiator", "thermostat", "filter", "pump", "valve",
    "bell", "whistle", "horn", "siren", "buzzer", "speaker", "headphone", "antenna",

    # ─── Places & Locations ───
    "house", "home", "apartment", "mansion", "cabin", "cottage", "villa", "palace",
    "castle", "fortress", "tower", "temple", "church", "mosque", "cathedral", "shrine",
    "school", "university", "college", "academy", "institute", "library", "museum", "gallery",
    "hospital", "clinic", "pharmacy", "laboratory", "office", "factory", "warehouse", "workshop",
    "store", "shop", "market", "mall", "bazaar", "boutique", "supermarket", "bakery",
    "restaurant", "cafe", "bar", "pub", "inn", "hotel", "motel", "resort",
    "park", "playground", "stadium", "arena", "theater", "cinema", "auditorium", "coliseum",
    "airport", "station", "port", "harbor", "dock", "pier", "terminal", "platform",
    "bridge", "tunnel", "highway", "road", "street", "avenue", "lane", "alley",
    "city", "town", "village", "suburb", "neighborhood", "district", "county", "state",
    "country", "nation", "kingdom", "empire", "republic", "territory", "province", "region",
    "farm", "ranch", "plantation", "vineyard", "greenhouse", "barn", "silo", "stable",
    "prison", "jail", "dungeon", "camp", "base", "outpost", "bunker", "shelter",
    "cemetery", "tomb", "grave", "memorial", "monument", "statue", "fountain", "plaza",

    # ─── Food & Drink ───
    "bread", "rice", "pasta", "noodle", "cereal", "oatmeal", "pancake", "waffle",
    "meat", "beef", "pork", "lamb", "chicken", "turkey", "bacon", "sausage",
    "fish", "shrimp", "crab", "lobster", "oyster", "clam", "mussel", "scallop",
    "egg", "cheese", "butter", "cream", "milk", "yogurt", "custard", "pudding",
    "apple", "banana", "orange", "grape", "strawberry", "blueberry", "raspberry", "cherry",
    "peach", "pear", "plum", "mango", "pineapple", "coconut", "watermelon", "melon",
    "lemon", "lime", "grapefruit", "kiwi", "avocado", "tomato", "potato", "onion",
    "garlic", "pepper", "chili", "carrot", "broccoli", "spinach", "lettuce", "cabbage",
    "corn", "bean", "pea", "lentil", "chickpea", "soybean", "tofu", "mushroom",
    "nut", "almond", "walnut", "cashew", "peanut", "pistachio", "hazelnut", "pecan",
    "sugar", "honey", "syrup", "chocolate", "candy", "cookie", "cake", "pie",
    "ice", "cream", "gelato", "sorbet", "pastry", "donut", "muffin", "brownie",
    "coffee", "tea", "juice", "smoothie", "lemonade", "soda", "wine", "beer",
    "water", "sparkling", "mineral", "cocktail", "whiskey", "vodka", "rum", "brandy",
    "salt", "pepper", "spice", "herb", "basil", "oregano", "thyme", "rosemary",
    "cinnamon", "ginger", "turmeric", "cumin", "paprika", "mustard", "ketchup", "sauce",
    "soup", "stew", "broth", "chowder", "salad", "sandwich", "burger", "pizza",
    "taco", "burrito", "sushi", "ramen", "curry", "steak", "roast", "grill",

    # ─── Body & Health ───
    "head", "face", "forehead", "temple", "cheek", "chin", "jaw", "skull",
    "eye", "ear", "nose", "mouth", "lip", "tongue", "tooth", "throat",
    "neck", "shoulder", "arm", "elbow", "wrist", "hand", "finger", "thumb",
    "chest", "rib", "back", "spine", "waist", "hip", "pelvis", "abdomen",
    "leg", "thigh", "knee", "shin", "ankle", "foot", "toe", "heel",
    "skin", "bone", "muscle", "tendon", "ligament", "cartilage", "joint", "nerve",
    "brain", "heart", "lung", "liver", "kidney", "stomach", "intestine", "bladder",
    "blood", "vein", "artery", "pulse", "breath", "oxygen", "cell", "tissue",
    "hair", "nail", "eyebrow", "eyelash", "beard", "mustache", "freckle", "wrinkle",
    "health", "fitness", "strength", "stamina", "endurance", "flexibility", "balance", "posture",
    "disease", "illness", "infection", "virus", "bacteria", "fever", "cough", "cold",
    "pain", "ache", "cramp", "spasm", "wound", "bruise", "scar", "blister",
    "medicine", "drug", "pill", "tablet", "capsule", "injection", "vaccine", "antidote",
    "surgery", "therapy", "treatment", "diagnosis", "symptom", "remedy", "prescription", "dosage",

    # ─── Relationships & People ───
    "friend", "enemy", "stranger", "neighbor", "companion", "partner", "colleague", "ally",
    "family", "parent", "mother", "father", "sister", "brother", "daughter", "son",
    "husband", "wife", "spouse", "couple", "marriage", "wedding", "divorce", "engagement",
    "baby", "child", "toddler", "teenager", "adult", "elder", "ancestor", "descendant",
    "uncle", "aunt", "cousin", "nephew", "niece", "grandparent", "grandmother", "grandfather",
    "teacher", "student", "mentor", "apprentice", "coach", "player", "captain", "leader",
    "boss", "employee", "worker", "manager", "director", "executive", "president", "founder",
    "doctor", "nurse", "patient", "surgeon", "dentist", "therapist", "pharmacist", "paramedic",
    "king", "queen", "prince", "princess", "knight", "warrior", "soldier", "general",
    "hero", "villain", "victim", "witness", "judge", "lawyer", "detective", "officer",
    "artist", "musician", "singer", "dancer", "actor", "writer", "poet", "author",
    "scientist", "engineer", "programmer", "designer", "architect", "inventor", "researcher", "professor",
    "farmer", "baker", "chef", "pilot", "sailor", "driver", "mechanic", "electrician",
    "merchant", "trader", "banker", "accountant", "journalist", "reporter", "editor", "photographer",

    # ─── Time & Temporal ───
    "second", "minute", "hour", "day", "week", "month", "year", "decade",
    "century", "millennium", "era", "epoch", "age", "period", "moment", "instant",
    "morning", "afternoon", "evening", "night", "midnight", "noon", "daybreak", "nightfall",
    "today", "tomorrow", "yesterday", "future", "past", "present", "forever", "never",
    "early", "late", "soon", "always", "often", "rarely", "sometimes", "frequently",
    "beginning", "middle", "ending", "origin", "conclusion", "deadline", "milestone", "anniversary",
    "schedule", "timeline", "calendar", "clock", "watch", "countdown", "duration", "interval",
    "history", "legacy", "tradition", "heritage", "memory", "flashback", "prophecy", "destiny",

    # ─── Science & Technology ───
    "atom", "molecule", "element", "compound", "reaction", "catalyst", "enzyme", "protein",
    "electron", "proton", "neutron", "photon", "quark", "ion", "isotope", "nucleus",
    "gravity", "force", "energy", "power", "voltage", "current", "frequency", "wavelength",
    "light", "dark", "spectrum", "radiation", "laser", "plasma", "magnetism", "electricity",
    "speed", "velocity", "acceleration", "momentum", "friction", "density", "mass", "weight",
    "temperature", "heat", "cold", "entropy", "thermodynamics", "kinetic", "potential", "mechanical",
    "biology", "chemistry", "physics", "geology", "astronomy", "ecology", "genetics", "evolution",
    "gene", "chromosome", "mutation", "adaptation", "selection", "species", "organism", "ecosystem",
    "computer", "algorithm", "software", "hardware", "network", "internet", "server", "database",
    "code", "program", "script", "function", "variable", "loop", "array", "object",
    "robot", "drone", "satellite", "rocket", "spacecraft", "telescope", "radar", "sonar",
    "experiment", "hypothesis", "theory", "evidence", "proof", "data", "research", "discovery",
    "invention", "innovation", "technology", "engineering", "manufacture", "automation", "artificial", "intelligence",
    "digital", "analog", "binary", "quantum", "virtual", "reality", "simulation", "model",

    # ─── Materials & Substances ───
    "wood", "metal", "steel", "iron", "copper", "gold", "silver", "bronze",
    "aluminum", "titanium", "platinum", "zinc", "tin", "lead", "nickel", "chrome",
    "plastic", "rubber", "silicon", "carbon", "graphite", "fiberglass", "ceramic", "porcelain",
    "concrete", "cement", "brick", "marble", "granite", "slate", "limestone", "sandstone",
    "fabric", "cotton", "silk", "wool", "linen", "polyester", "nylon", "velvet",
    "leather", "suede", "fur", "feather", "ivory", "bone", "shell", "pearl",
    "paper", "cardboard", "foam", "wax", "resin", "glue", "tape", "adhesive",
    "oil", "fuel", "gasoline", "diesel", "coal", "charcoal", "propane", "kerosene",

    # ─── Clothing & Accessories ───
    "shirt", "pants", "jeans", "shorts", "skirt", "dress", "suit", "jacket",
    "coat", "sweater", "hoodie", "vest", "blazer", "uniform", "costume", "robe",
    "hat", "cap", "helmet", "hood", "scarf", "glove", "mitten", "sock",
    "shoe", "boot", "sandal", "slipper", "sneaker", "heel", "loafer", "moccasin",
    "belt", "tie", "bow", "button", "zipper", "buckle", "clasp", "strap",
    "glasses", "sunglasses", "goggles", "mask", "veil", "headband", "bandana", "turban",

    # ─── Transportation ───
    "car", "truck", "bus", "van", "taxi", "ambulance", "motorcycle", "bicycle",
    "train", "subway", "tram", "trolley", "cable", "monorail", "locomotive", "wagon",
    "airplane", "helicopter", "jet", "glider", "blimp", "balloon", "parachute", "drone",
    "boat", "ship", "yacht", "canoe", "kayak", "raft", "ferry", "submarine",
    "skateboard", "scooter", "rollerblades", "wheelchair", "cart", "sled", "carriage", "chariot",
    "engine", "wheel", "tire", "brake", "steering", "pedal", "gear", "transmission",

    # ─── Sports & Games ───
    "soccer", "football", "basketball", "baseball", "tennis", "volleyball", "hockey", "golf",
    "cricket", "rugby", "boxing", "wrestling", "martial", "karate", "judo", "fencing",
    "swimming", "diving", "surfing", "skiing", "snowboarding", "skating", "cycling", "rowing",
    "running", "jogging", "sprinting", "marathon", "relay", "hurdle", "triathlon", "decathlon",
    "archery", "shooting", "hunting", "fishing", "climbing", "hiking", "camping", "trekking",
    "chess", "checkers", "poker", "domino", "puzzle", "riddle", "maze", "trivia",
    "ball", "goal", "score", "point", "match", "game", "tournament", "championship",
    "team", "opponent", "referee", "coach", "trophy", "medal", "record", "victory",

    # ─── Music & Sound ───
    "music", "song", "melody", "harmony", "rhythm", "beat", "tempo", "pitch",
    "note", "chord", "scale", "key", "tune", "lyric", "verse", "chorus",
    "guitar", "piano", "violin", "drum", "flute", "trumpet", "saxophone", "harp",
    "bass", "cello", "clarinet", "oboe", "trombone", "tuba", "harmonica", "accordion",
    "singer", "band", "orchestra", "choir", "concert", "recital", "performance", "festival",
    "volume", "sound", "noise", "silence", "echo", "vibration", "resonance", "frequency",
    "rock", "pop", "jazz", "blues", "classical", "country", "folk", "reggae",
    "rap", "hip", "electronic", "techno", "ambient", "soul", "funk", "gospel",

    # ─── Abstract Concepts ───
    "truth", "lie", "fact", "fiction", "reality", "illusion", "dream", "nightmare",
    "life", "death", "birth", "rebirth", "soul", "spirit", "mind", "consciousness",
    "freedom", "justice", "equality", "liberty", "democracy", "tyranny", "oppression", "revolution",
    "peace", "war", "conflict", "harmony", "chaos", "order", "balance", "stability",
    "good", "evil", "right", "wrong", "moral", "virtue", "vice", "ethics",
    "beauty", "ugliness", "elegance", "grace", "charm", "allure", "splendor", "magnificence",
    "wisdom", "knowledge", "ignorance", "intelligence", "genius", "talent", "skill", "ability",
    "power", "authority", "influence", "control", "dominance", "submission", "obedience", "rebellion",
    "success", "failure", "achievement", "accomplishment", "progress", "decline", "growth", "stagnation",
    "wealth", "poverty", "abundance", "scarcity", "luxury", "necessity", "comfort", "hardship",
    "danger", "safety", "risk", "security", "threat", "protection", "vulnerability", "resilience",
    "change", "transformation", "transition", "evolution", "revolution", "innovation", "tradition", "custom",
    "purpose", "meaning", "significance", "value", "importance", "priority", "relevance", "impact",
    "mystery", "secret", "puzzle", "enigma", "paradox", "irony", "coincidence", "fate",
    "imagination", "creativity", "inspiration", "vision", "intuition", "instinct", "perception", "awareness",
    "language", "communication", "expression", "gesture", "symbol", "metaphor", "analogy", "allegory",

    # ─── Colors ───
    "red", "blue", "green", "yellow", "orange", "purple", "pink", "brown",
    "black", "white", "gray", "silver", "gold", "bronze", "copper", "ivory",
    "crimson", "scarlet", "maroon", "burgundy", "coral", "salmon", "magenta", "fuchsia",
    "cyan", "teal", "turquoise", "aqua", "navy", "indigo", "cobalt", "azure",
    "lime", "olive", "emerald", "jade", "mint", "sage", "forest", "chartreuse",
    "amber", "golden", "tan", "beige", "khaki", "cream", "peach", "apricot",
    "violet", "lavender", "lilac", "plum", "mauve", "orchid", "periwinkle", "amethyst",

    # ─── Math & Numbers ───
    "number", "digit", "integer", "fraction", "decimal", "percent", "ratio", "proportion",
    "addition", "subtraction", "multiplication", "division", "equation", "formula", "calculation", "computation",
    "geometry", "algebra", "calculus", "statistics", "probability", "logic", "theorem", "axiom",
    "circle", "square", "triangle", "rectangle", "sphere", "cube", "cylinder", "cone",
    "angle", "radius", "diameter", "circumference", "area", "volume", "perimeter", "surface",
    "zero", "one", "two", "three", "four", "five", "six", "seven",
    "eight", "nine", "ten", "hundred", "thousand", "million", "billion", "infinity",
    "half", "double", "triple", "quadruple", "single", "pair", "dozen", "score",

    # ─── Society & Culture ───
    "law", "rule", "regulation", "policy", "constitution", "amendment", "statute", "ordinance",
    "government", "parliament", "congress", "senate", "council", "committee", "cabinet", "ministry",
    "election", "vote", "ballot", "campaign", "debate", "rally", "protest", "demonstration",
    "culture", "civilization", "society", "community", "population", "citizen", "immigrant", "refugee",
    "religion", "faith", "belief", "worship", "prayer", "meditation", "ritual", "ceremony",
    "festival", "celebration", "holiday", "tradition", "custom", "folklore", "myth", "legend",
    "art", "literature", "poetry", "drama", "comedy", "tragedy", "fiction", "novel",
    "education", "curriculum", "diploma", "degree", "scholarship", "thesis", "dissertation", "lecture",

    # ─── Architecture & Construction ───
    "building", "structure", "foundation", "frame", "column", "beam", "arch", "dome",
    "window", "door", "gate", "fence", "wall", "barrier", "railing", "balcony",
    "room", "hall", "lobby", "corridor", "passage", "attic", "basement", "cellar",
    "kitchen", "bathroom", "bedroom", "living", "dining", "study", "closet", "pantry",
    "floor", "ceiling", "roof", "chimney", "fireplace", "porch", "deck", "patio",
    "garden", "lawn", "yard", "driveway", "sidewalk", "pathway", "trail", "road",

    # ─── Shapes & Patterns ───
    "line", "curve", "wave", "zigzag", "spiral", "helix", "loop", "arc",
    "dot", "spot", "stripe", "band", "ring", "oval", "diamond", "star",
    "grid", "lattice", "mesh", "web", "pattern", "design", "symmetry", "asymmetry",
    "texture", "smooth", "rough", "bumpy", "flat", "round", "sharp", "blunt",

    # ─── Communication & Media ───
    "word", "sentence", "paragraph", "chapter", "page", "document", "article", "essay",
    "speech", "address", "announcement", "declaration", "statement", "testimony", "confession", "apology",
    "message", "email", "text", "chat", "call", "signal", "broadcast", "transmission",
    "newspaper", "magazine", "blog", "website", "podcast", "radio", "television", "media",
    "story", "narrative", "plot", "character", "setting", "conflict", "resolution", "climax",
    "question", "answer", "response", "reply", "comment", "opinion", "argument", "debate",

    # ─── Warfare & Conflict ───
    "sword", "shield", "armor", "helmet", "spear", "arrow", "bow", "crossbow",
    "gun", "rifle", "pistol", "cannon", "bomb", "missile", "grenade", "mine",
    "tank", "fighter", "bomber", "warship", "battleship", "destroyer", "cruiser", "carrier",
    "army", "navy", "force", "troop", "regiment", "battalion", "brigade", "division",
    "strategy", "tactics", "formation", "maneuver", "siege", "retreat", "advance", "flank",
    "treaty", "alliance", "ceasefire", "armistice", "surrender", "occupation", "liberation", "independence",

    # ─── Qualities & Descriptors ───
    "big", "small", "large", "tiny", "huge", "massive", "enormous", "miniature",
    "tall", "short", "long", "wide", "narrow", "thick", "thin", "slim",
    "fast", "slow", "quick", "rapid", "swift", "gradual", "steady", "constant",
    "hot", "warm", "cool", "cold", "freezing", "boiling", "lukewarm", "mild",
    "hard", "soft", "tough", "gentle", "strong", "weak", "fragile", "sturdy",
    "old", "new", "young", "ancient", "modern", "fresh", "stale", "vintage",
    "bright", "dim", "vivid", "dull", "shiny", "matte", "glossy", "opaque",
    "loud", "quiet", "silent", "noisy", "peaceful", "chaotic", "turbulent", "calm",
    "clean", "dirty", "pure", "contaminated", "spotless", "filthy", "tidy", "messy",
    "wet", "dry", "moist", "damp", "soaked", "parched", "humid", "arid",
    "full", "empty", "packed", "sparse", "crowded", "vacant", "dense", "hollow",
    "rich", "poor", "expensive", "cheap", "valuable", "worthless", "priceless", "affordable",
    "simple", "complex", "easy", "difficult", "basic", "advanced", "elementary", "sophisticated",
    "true", "false", "real", "fake", "genuine", "artificial", "authentic", "counterfeit",
    "safe", "dangerous", "secure", "risky", "stable", "volatile", "reliable", "unpredictable",
    "open", "closed", "public", "private", "visible", "hidden", "transparent", "opaque",
    "alive", "dead", "active", "dormant", "awake", "asleep", "conscious", "unconscious",
    "raw", "cooked", "ripe", "unripe", "mature", "immature", "developed", "primitive",

    # ─── Miscellaneous / Fill ───
    "fire", "water", "earth", "wind", "air", "metal", "nature", "time",
    "space", "void", "portal", "gateway", "threshold", "boundary", "border", "edge",
    "center", "middle", "core", "heart", "focus", "target", "goal", "aim",
    "source", "origin", "root", "foundation", "base", "pillar", "cornerstone", "keystone",
    "path", "route", "journey", "adventure", "quest", "mission", "expedition", "voyage",
    "home", "haven", "sanctuary", "refuge", "asylum", "retreat", "paradise", "utopia",
    "shadow", "phantom", "ghost", "spirit", "specter", "apparition", "mirage", "echo",
    "flame", "ember", "spark", "blaze", "inferno", "bonfire", "campfire", "wildfire",
    "storm", "tempest", "maelstrom", "vortex", "whirlwind", "tornado", "cyclone", "hurricane",
    "treasure", "fortune", "bounty", "reward", "prize", "gift", "offering", "tribute",
    "magic", "spell", "enchantment", "charm", "curse", "hex", "potion", "elixir",
    "story", "tale", "fable", "myth", "legend", "saga", "epic", "chronicle",
    "crown", "throne", "scepter", "orb", "crest", "seal", "emblem", "insignia",
    "riddle", "clue", "hint", "puzzle", "mystery", "secret", "code", "cipher",
    "oath", "vow", "pledge", "promise", "bond", "pact", "covenant", "contract",
    "trial", "test", "challenge", "ordeal", "competition", "contest", "race", "duel",
    "dawn", "dusk", "twilight", "midnight", "eclipse", "aurora", "constellation", "meteor",
    "crystal", "prism", "lens", "mirror", "reflection", "refraction", "spectrum", "wavelength",
    "anchor", "compass", "map", "chart", "atlas", "globe", "horizon", "landmark",
    "mask", "disguise", "camouflage", "costume", "cloak", "shroud", "veil", "curtain",
    "bridge", "arch", "tower", "lighthouse", "beacon", "watchtower", "monument", "obelisk",
    "garden", "orchard", "vineyard", "greenhouse", "nursery", "plantation", "field", "meadow",
    "market", "bazaar", "auction", "exchange", "commerce", "trade", "merchant", "vendor",
    "kingdom", "empire", "dynasty", "realm", "domain", "territory", "dominion", "province",
    "citizen", "peasant", "noble", "knight", "samurai", "gladiator", "viking", "pirate",
    "dragon", "phoenix", "unicorn", "griffin", "centaur", "mermaid", "werewolf", "vampire",
    "wizard", "witch", "sorcerer", "shaman", "druid", "prophet", "oracle", "sage",
    "alchemy", "potion", "elixir", "talisman", "amulet", "relic", "artifact", "idol",
]))


def main():
    output_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "data")
    os.makedirs(output_dir, exist_ok=True)

    vocab_path = os.path.join(output_dir, "vocabulary.json")
    embeddings_path = os.path.join(output_dir, "embeddings.npy")

    vocab = list(VOCABULARY)
    print(f"[precompute] Vocabulary size: {len(vocab)} words")
    print(f"[precompute] Loading model: all-MiniLM-L6-v2 ...")

    t0 = time.time()
    model = SentenceTransformer("all-MiniLM-L6-v2")
    print(f"[precompute] Model loaded in {time.time() - t0:.1f}s")

    # Embed in batches
    batch_size = 256
    print(f"[precompute] Encoding {len(vocab)} words in batches of {batch_size} ...")

    t1 = time.time()
    embeddings = model.encode(vocab, batch_size=batch_size, show_progress_bar=True, normalize_embeddings=True)
    embeddings = np.array(embeddings, dtype=np.float32)
    print(f"[precompute] Encoding complete in {time.time() - t1:.1f}s")
    print(f"[precompute] Embeddings shape: {embeddings.shape}")

    # Save vocabulary
    with open(vocab_path, "w", encoding="utf-8") as f:
        json.dump(vocab, f)
    print(f"[precompute] Saved vocabulary → {vocab_path}")

    # Save embeddings
    np.save(embeddings_path, embeddings)
    print(f"[precompute] Saved embeddings → {embeddings_path}")
    print(f"[precompute] Total time: {time.time() - t0:.1f}s")
    print(f"[precompute] Done! You can now run the Go server without Python.")


if __name__ == "__main__":
    main()
