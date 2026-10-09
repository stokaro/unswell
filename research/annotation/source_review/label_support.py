"""Count observed supervision without treating missing labels as negatives."""

from annotation_contract import INTENTS


def summarize_label_support(reviews):
    """Summarize normalized source reviews; class support is not qualification.

    Each mapping key identifies one reviewed source case. Optional or unknown
    decisions count as unobserved. This function does not infer token labels,
    grade reviews, choose thresholds, or declare a model available.
    """
    if not isinstance(reviews, dict):
        raise ValueError("Reviews must be a mapping of source identities")
    names = ("necessity",) + INTENTS
    counts = {name: {"positive": 0, "negative": 0, "unobserved": 0} for name in names}
    for identity, review in reviews.items():
        if not isinstance(identity, str) or not identity or not isinstance(review, dict):
            raise ValueError("A review needs a nonempty source identity and record")
        necessity = review.get("necessity")
        labels = review.get("intent_loss_labels")
        masks = review.get("intent_loss_mask")
        if (not isinstance(necessity, dict) or "loss_label" not in necessity
                or not isinstance(labels, dict) or set(labels) != set(INTENTS)
                or not isinstance(masks, dict) or set(masks) != set(INTENTS)):
            raise ValueError("A normalized review needs every label and explicit mask")
        values = {"necessity": necessity["loss_label"], **labels}
        for name, value in values.items():
            if value is not None and (type(value) is not int or value not in (0, 1)):
                raise ValueError("Observed labels must be binary integers")
            if name != "necessity" and (type(masks[name]) is not bool
                                        or masks[name] != (value is not None)):
                raise ValueError("Intent label and observation mask disagree")
            selected = "unobserved" if value is None else "positive" if value else "negative"
            counts[name][selected] += 1
    for row in counts.values():
        row["both_classes_observed"] = row["positive"] > 0 and row["negative"] > 0
    return {"source_cases": len(reviews), "labels": counts,
            "class_support_is_model_qualification": False,
            "unmarked_tokens_are_background": False}
