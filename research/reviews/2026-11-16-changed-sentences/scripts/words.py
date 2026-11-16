"""The words by which `unswell compare` compares two sentences, as review/sentences.go reads them.

content_words(text) returns the content words of a text without their common endings. It is here so that a rule
of the final build can be applied to answers that an earlier build of the fitting half kept.
"""
import re

NEGATIONS = set("not no never none nor neither without cannot nothing nobody nowhere".split())
FUNCTION = set("""a an the and or but if then else so as at by for from in into of on onto to with over under about
after before between through during until while when where which who whom whose what why how that this these those
there here it its they them their we us our you your he she his her i me my is are was were be been being am
do does did done doing have has had having will would shall should can could may might must also just very more
most much many some any each every all both either than too only own same such other another one two up down out
off again further once because since though although however yet still even now""".split())
ENDINGS = (("ies", "y", 4), ("ied", "y", 4), ("ing", "", 5), ("ed", "", 4))


def stem(word):
    for ending, instead, longer_than in ENDINGS:
        if len(word) > longer_than and word.endswith(ending):
            word = word[:-len(ending)] + instead
            break
    else:
        if len(word) > 3 and word.endswith("s") and not word.endswith("ss"):
            word = word[:-1]
    if len(word) > 3 and word.endswith("e"):
        word = word[:-1]
    return word


def content_words(text):
    words = set()
    for token in re.split(r"[^\w'’]+", text.lower().replace("_", " ")):
        token = token.replace("’", "'").strip("'")
        if not token or token in NEGATIONS or token.endswith("n't"):
            continue
        token = token.split("'")[0]
        if token and token not in FUNCTION:
            words.add(stem(token))
    return words
