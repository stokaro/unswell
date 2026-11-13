"""Say whether a commit message states that it removes AI-sounding wording."""
import re

AI = r"(?:ai|llm|gpt|chatgpt|claude|copilot|robot|machine)"
STATES = re.compile("|".join([
    r"\bde-?slop", r"\bun-?slop", rf"\b{AI}[- ]?slop\b", rf"\bslop\b[^\n]{{0,40}}\b(readme|docs?|wording|prose|copy|text|writing)\b",
    rf"\b{AI}[- ]?(?:isms?|fluff|tells?|speak|voice|sounding|smell|vibes?|cliches?|tics?|tropes?|wording|tone|phrasing|prose)\b",
    rf"\bsounds? (?:like|too|less|so) (?:an? |much like |written by )?{AI}\b", rf"\bless {AI}\b", rf"\bun-?{AI}\b",
    rf"\b(?:remove|strip|drop|cut|fix|clean\w*|reduce|tone down|rewrite|kill|purge|trim)\b[^\n]{{0,50}}\b{AI}[- ](?:generated|written|sounding|style|ish|like)\b",
    r"\bhuman(?:ize|ise)[sd]?\b[^\n]{0,40}\b(readme|docs?|copy|wording|prose|text|writing|tone|language)\b",
]), re.I)
ACTOR = re.compile(rf"\b(?:generated|written|rewritten|created|drafted|improved|polished|updated)\b[^\n]{{0,15}}\b(?:by|with|using|via)\b[^\n]{{0,10}}\b{AI}\b", re.I)


def strict(message):
    first = message.strip().split("\n\n")[0]
    return bool(STATES.search(first)) and not (ACTOR.search(first) and not re.search(r"de-?slop|slop|isms?|sounding|tells", first, re.I))
