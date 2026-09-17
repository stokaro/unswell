# Reviewed changes

Indices refer to the retained strict reports. Replaced findings link to their after review; exact before and after judgments remain in dispositions.json.

## development / added

### 101: filler.instruction-scaffolding

The need-to purpose frame delays the named output post-processing hook. This instruction can be stated with a direct purpose clause while retaining the command-specific hook. It was not frozen.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: To post-process fzf output, define _fzf_complete_COMMAND_post as shown.

`sources/p019.md` bytes 15191–15286

```text
If you need to post-process the output from fzf, define
`_fzf_complete_COMMAND_post` as follows
```

## exposed_framing / added

### 114: filler.instruction-scaffolding

The reader-desire and can-use frame repeats the preview goal before the actual command. The custom file and fzf prerequisite remain necessary. This was outside the earlier frozen category scope.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: To preview themes on a custom file, run the command; retain the fzf prerequisite.

`sources/c09.md` bytes 10282–10438

```text
If you want to preview the different themes on a custom file, you can use
the following command (you need [`fzf`](https://github.com/junegunn/fzf) for this)
```

### 115: filler.instruction-scaffolding

The reader-desire frame can become a purpose clause; both variables and the override ordering must remain. This was not frozen.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: To use another pager, modify PAGER or set BAT_PAGER to override it.

`sources/c09.md` bytes 13959–14124

```text
If you want to use a different pager, you can either modify the
`PAGER` variable or set the `BAT_PAGER` environment variable to override what is specified in
`PAGER`
```

### 116: filler.instruction-scaffolding

The reader-desire/can-pass sequence delays the option and its older-version condition. The mouse-wheel versus quit-if-one-screen tradeoff must remain. This was not frozen.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: For mouse-wheel scrolling on older less versions, pass -R; retain the disabled quit-if-one-screen feature and the separate 530-or-newer behavior.

`sources/c09.md` bytes 15239–15410

```text
If you want to enable mouse-wheel scrolling on older versions of `less`, you can pass just `-R` (as
in the example above, this will disable the quit-if-one-screen feature)
```

## exposed_instruction / added

### 137: filler.instruction-scaffolding

The reader-desire frame precedes the actual invocation. Its two goals and possible forwarding of received cookies remain necessary; this wording repair was not frozen.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: To accept received cookies and follow redirects, invoke curl as shown; retain the possible forwarding of received cookies.

`sources/c08.md` bytes 21107–21259

```text
if you want to let curl understand cookies from a
 page and follow a location (and thus possibly send back cookies it received),
 you can invoke it like
```

### 168: filler.instruction-scaffolding

The can-be-used-to construction could be shorter, but the frozen review left this aggregation introduction uncertain because the input, output and dimensionality matter. Keep it unresolved.

Status: uncertain. Full: none. Partial: none.

`sources/c10.md` bytes 7971–8173

```text
Prometheus supports the following built-in aggregation operators that can be
used to aggregate the elements of a single instant vector, resulting in a new
vector of fewer elements with aggregated values
```

### 169: filler.instruction-scaffolding

The reader-desire frame delays the indentation action; the preserved line breaks, tab and four-space alternative remain in scope.

Status: actionable. Full: c11-d03. Partial: none.

`sources/c11.md` bytes 982–1090

```text
If you want to include code and have new
lines preserved, indent the line with a tab
or at least four spaces
```

## exposed_purpose / added

### 18: filler.instruction-scaffolding

The relative support chain is diagnosed, but the repeated informational-label/additional-information wording is not specifically identified. Award partial coverage of the compound defect only.

Status: actionable. Full: none. Partial: c06-d02.

`sources/c06.md` bytes 1393–1557

```text
The `annotations` clause specifies a set of informational labels that can be used to store longer additional information such as alert descriptions or runbook links
```

## exposed_relations / added

### 21: filler.instruction-scaffolding

The allows-reader/specify/to-be-used layers diagnose the frozen referrer wrapper while retaining its command-line location.

Status: actionable. Full: c04-d08. Partial: none.

`sources/c04.md` bytes 12828–12898

```text
Curl allows you to specify the referrer to be
used on the command line
```

### 24: filler.instruction-scaffolding

The generic capability setup is related to the explicit switch-based method through the transfer-speed object. The warning includes both locations and preserves the limit and duration.

Status: actionable. Full: c04-d14. Partial: none.

`sources/c04.md` bytes 17901–18006

```text
Curl allows the user to set the transfer speed conditions that must be met to
let the transfer keep going
```

`sources/c04.md` bytes 18008–18154

```text
By using the switch `-y` and `-Y` you can make
curl abort transfers if the transfer speed is below the specified lowest limit
for a specified time
```

### 27: filler.instruction-scaffolding

Capable-of-using can become can-use while retaining client certificates, the remote requirement and the subsequent PEM constraint. This repair was not frozen.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: Curl can also use client certificates to get or post files from sites that require valid certificates; keep the PEM requirement.

`sources/c04.md` bytes 24163–24273

```text
curl is also capable of using client certificates to get/post files from sites
that require valid certificates
```

## exposed_proposition / added

### 12: filler.instruction-scaffolding

The checksum is an operand used for caching, not an actor that itself caches. The relative clause explains its role in the dependency-derived cache key; an action-support diagnosis does not establish a safe direct action here.

Status: nonactionable. Full: none. Partial: none.

`sources/c04.md` bytes 6338–6489

```text
It contains a digest that is combined with the cache keys of the inputs to determine the stable checksum that can be used to cache the operation result
```

### 21: filler.instruction-scaffolding

The editor directory and configuration consequence are a frozen repetition control and must remain. The separate imperative be-sure-to wrapper can be shortened without losing that condition; it was not a frozen defect.

Status: additional_actionable. Full: none. Partial: none.

Additional repair: When using Visual Studio Code, open web/ui/react-app instead of the repository root; retain the ESLint and TypeScript reason.

`sources/c06.md` bytes 1360–1460

```text
be sure to open the `web/ui/react-app` directory in the editor instead of the root of the repository
```

## exposed_clauses / added

### 34: filler.instruction-scaffolding

The relative can-be-used-to/apply-parsing-logic chain identifies the frozen indirect API description. Keep both method identities and the subsequent distinction about arbitrary compatible types.

Status: actionable. Full: c04-d19. Partial: none.

`sources/c04.md` bytes 21478–21638

```text
Pydantic includes a standalone utility function `parse_obj_as` that can be used to apply the parsing
logic used to populate pydantic models in a more ad-hoc way
```

### 35: filler.instruction-scaffolding

The capable-of-parsing layer is exactly the frozen wording defect. Can parse preserves the same supported field-type range.

Status: actionable. Full: c04-d20. Partial: none.

`sources/c04.md` bytes 21977–22086

```text
This function is capable of parsing data into any of the types pydantic can handle as fields of a `BaseModel`
```

### 37: filler.instruction-scaffolding

The imperative assurance wrapper delays the sign-off requirement. Retain the DCO link and requiredness.

Status: actionable. Full: c06-d02. Partial: none.

`sources/c06.md` bytes 1111–1142

```text
Be sure to sign off on the [DCO
```

## confirmation / added

### 21: filler.instruction-scaffolding

The reader-intention and replacement layers wrap the exact allocator override instruction. The warning retains allocator examples and the separate pre-allocation ordering requirement.

Status: actionable. Full: c04-d09. Partial: none.

`sources/c04.txt` bytes 5923–6131

```text
If you want to override the allocation functions used by libevent
  (for example, to use a specialized allocator, or debug memory
  issues, or so on), you can replace them by calling
  event_set_mem_functions
```

### 33: filler.instruction-scaffolding

The diagnostic relates both SSL reader-desire wrappers. The filter-based and direct-socket alternatives remain separate and both frozen events are covered.

Status: actionable. Full: c04-d25, c04-d26. Partial: none.

`sources/c04.txt` bytes 21514–21645

```text
If you want to wrap an
   SSL layer around an existing bufferevent, you would call the
   bufferevent_openssl_filter_new() function
```

`sources/c04.txt` bytes 21648–21732

```text
If you want to do SSL
   on a socket directly, call bufferevent_openssl_socket_new()
```

### 34: filler.instruction-scaffolding

The need-to-limit/can-do-this support chain wraps the same rate-limiting operation. Keep byte direction and single-versus-group scope.

Status: actionable. Full: c04-d28. Partial: none.

`sources/c04.txt` bytes 22529–22702

```text
If you need to limit the number of bytes read/written by a single
   bufferevent, or by a group of them, you can do this with a new set of
   bufferevent rate-limiting calls
```
