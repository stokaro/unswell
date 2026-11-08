"""Verify the published result from frozen inputs and complete raw run evidence."""

import packet
import run
import score


def verify():
    run.verify()
    actual = score.score()
    if actual != packet.read(packet.ROOT / "result.json"):
        raise ValueError("published pilot result differs from complete raw evidence")
    return actual


if __name__ == "__main__":
    result = verify()
    print("source-bound disposition pilot:", result["arms"]["treatment"]["full"],
          "FULL; advance", result["advance"])
