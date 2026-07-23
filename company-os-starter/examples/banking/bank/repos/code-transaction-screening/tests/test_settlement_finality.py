"""Consumer-side conformance tests for the screening service.

Code repos are NOT modeled by the Company OS CLI. The only binding back to
governance is grep-able @spec markers tying tests to EARS clauses.
"""


# @spec req://payments/settlement-finality@1.0#R2
def test_duplicate_payment_order_is_screened_once():
    """R2: a duplicated Payment Order event must not produce a second Alert."""
    ...


# @spec req://identity/token-verification@1.0#R1
def test_rejects_unverified_token():
    """R1: every inbound call without a verifiable token is rejected."""
    ...
