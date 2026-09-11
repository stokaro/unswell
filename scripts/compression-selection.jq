# Select the reference seeds of the compression experiment from a candidate
# artifact. $budget bounds the bytes of one reference, and $level names the
# compression level of the bank.
def take($budget): reduce .[] as $u ({ids: [], bytes: 0};
    if (.bytes + ($u.unit.text | utf8bytelength) + 1) <= $budget and (.ids | length) < 128
    then {ids: (.ids + [$u.unit.id]), bytes: (.bytes + ($u.unit.text | utf8bytelength) + 1)} else . end) | .ids;
def alternate($a; $b): [range(0; ([$a, $b] | map(length) | max))] | map([$a[.], $b[.]]) | flatten | map(select(. != null));
  [.units[] | select(.partition == "training" and .unit.kind == "paragraph")]
  | group_by(.group_id)
  | map({group: .[0].group_id,
         historical: (map(select(.cohort | startswith("historical"))) | sort_by(.unit.id)),
         contemporary: (map(select(.cohort == "contemporary")) | sort_by(.unit.id))})
  | map(select((.historical | length) > 0 and (.contemporary | length) > 0))
  | max_by([([(.historical | length), (.contemporary | length)] | min), .group])
  | . as $seeds
  | {version: "unswell-compression-bank-v1", kind: "paragraph", allow_simulation: false,
     compression: {level: $level, max_input_bytes: 1048576},
     cohorts: [
       {id: "historical", origin: "historical", unit_ids: ($seeds.historical | take($budget))},
       {id: "contemporary", origin: "contemporary", unit_ids: ($seeds.contemporary | take($budget))},
       {id: "mixed", origin: "mixed", unit_ids: (alternate($seeds.contemporary; $seeds.historical) | take($budget))}
     ]}
