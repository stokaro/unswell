# Diagnostic changes

Both profiles add one actionable diagnostic and remove twelve development
diagnostics: eleven nonactionable findings and one actionable finding.
Confirmation removes one nonactionable finding and adds none.

Every change below has source-bound review in `dispositions.json`. Unchanged
development findings inherit the original audit and framing follow-up judgments.
This is assistant review; no independent agreement statistic is claimed.
Rendered snippets trim trailing line whitespace; the JSON reports retain exact text.

## development: added repetition.duplicate-list-item (actionable)

Both complete entries describe the same Literal documentation change and #651, within release 0.31. The diagnostic locates exactly both frozen targets, including the protected identifier and unchanged issue reference.

Diagnostic index: 82.

sources/p014.md, bytes [28114, 28169):

```text
add documentation for `Literal` type, #651 by @dmontagu
```

sources/p014.md, bytes [28380, 28433):

```text
add documentation for Literal type, #651 by @dmontagu
```

## development: removed repetition.paragraph-openers (actionable)

Known regression: section scoping removes the independent exit-0 setup from this cluster. The two annotated exit-1 repetitions no longer exceed the unchanged allowed-occurrences count of two. The frozen actionable judgment p004-d01 is retained; its loss is counted.

Diagnostic index: 41.

sources/p004.md, bytes [3346, 3500):

```text
The command exits `0` whether or not differences exist; add `--exit-code` to
exit `1` on a non-empty diff when a script needs the answer as a status code.
```

sources/p004.md, bytes [3877, 3977):

```text
The command exits `1` because drift was found, which is the check working
rather than a fault in it.
```

sources/p004.md, bytes [4158, 4204):

```text
The command exits `1` because drift was found.
```

## development: removed repetition.sentence-openers (nonactionable)

These openings belong to separate sections describing different export properties. YAML/HCL/Go introduces formats, metadata coverage or attribute categories; the That is sentences refer to different GraphQL and stored-type claims. Similar openings are not the annotated defect.

Diagnostic index: 51.

sources/p007.mdx, bytes [6146, 6213):

```text
YAML, HCL, and Go can also carry the API metadata documented below.
```

sources/p007.mdx, bytes [11944, 12017):

```text
YAML, HCL, and Go annotations can author every export-metadata attribute.
```

sources/p007.mdx, bytes [16729, 16852):

```text
YAML, HCL, and Go annotations provide one shared attribute and
three target-specific attributes on both tables and columns:
```

Diagnostic index: 53.

sources/p007.mdx, bytes [8749, 8821):

```text
That is the whole
default export: no `Query`, no inputs, no connections.
```

sources/p007.mdx, bytes [9278, 9377):

```text
That is the correct shape for a
type-system document meant to be composed into a schema you design.
```

sources/p007.mdx, bytes [20869, 20985):

```text
That is the right answer nearly always, and the wrong
one exactly where the stored type is not what the value means:
```

## development: removed repetition.exact-sentence (nonactionable)

The equal capability sentences occupy distinct provider rows. Each row is independently useful; removing the cross-cell comparison preserves both facts.

Diagnostic index: 56.

sources/p007.mdx, bytes [13034, 13103):

```text
YAML and HCL payloads carry API export metadata; SQL payloads do not.
```

sources/p007.mdx, bytes [13157, 13226):

```text
YAML and HCL payloads carry API export metadata; SQL payloads do not.
```

## development: removed repetition.near-sentence (nonactionable)

Sponsor credits belong to different release sections, with recipient sets that change. They are intentional release records rather than repeated change descriptions.

Diagnostic index: 86.

sources/p014.md, bytes [24, 213):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api,
@mkeen, @meadsteve for their kind support.
```

sources/p014.md, bytes [1233, 1410):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api, @mkeen
for their kind support.
```

sources/p014.md, bytes [1995, 2165):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api
for their kind support.
```

## development: removed repetition.paragraph-openers (nonactionable)

The same acknowledgment opening introduces credits for different releases and different recipients. Each credit belongs with its release.

Diagnostic index: 87.

sources/p014.md, bytes [24, 213):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api,
@mkeen, @meadsteve for their kind support.
```

sources/p014.md, bytes [1233, 1410):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api, @mkeen
for their kind support.
```

sources/p014.md, bytes [1995, 2165):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api
for their kind support.
```

sources/p014.md, bytes [6428, 6554):

```text
Thank you to pydantic's sponsors: @matin, @tiangolo, @chdsbd, @jorgecarleitao, and 1 anonymous sponsor for their kind support.
```

## development: removed repetition.sentence-openers (nonactionable)

Repeated acknowledgment syntax across releases does not make the distinct release credits redundant.

Diagnostic index: 88.

sources/p014.md, bytes [24, 213):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api,
@mkeen, @meadsteve for their kind support.
```

sources/p014.md, bytes [1233, 1410):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api, @mkeen
for their kind support.
```

sources/p014.md, bytes [1995, 2165):

```text
Thank you to pydantic's sponsors:
@timdrijvers, @BCarley, @chdsbd, @tiangolo, @matin, @linusg, @kevinalh, @jorgecarleitao, @koxudaxi, @primer-api
for their kind support.
```

sources/p014.md, bytes [6428, 6554):

```text
Thank you to pydantic's sponsors: @matin, @tiangolo, @chdsbd, @jorgecarleitao, and 1 anonymous sponsor for their kind support.
```

## development: removed repetition.paragraph-openers (nonactionable)

The conditional introductions name different operating systems, package managers and architectures, followed by different installation commands. Their common lookup format is useful.

Diagnostic index: 89.

sources/p017.md, bytes [10276, 10380):

```text
If you're a **macOS Homebrew** or a **Linuxbrew** user, then you can install
ripgrep from homebrew-core:
```

sources/p017.md, bytes [10414, 10559):

```text
If you're a **MacPorts** user, then you can install ripgrep from the
[official ports](https://www.macports.org/ports.php?by=name&substr=ripgrep):
```

sources/p017.md, bytes [10598, 10734):

```text
If you're a **Windows Chocolatey** user, then you can install ripgrep from the
[official repo](https://chocolatey.org/packages/ripgrep):
```

sources/p017.md, bytes [10769, 10933):

```text
If you're a **Windows Scoop** user, then you can install ripgrep from the
[official bucket](https://github.com/ScoopInstaller/Main/blob/master/bucket/ripgrep.json):
```

sources/p017.md, bytes [10968, 11055):

```text
If you're an **Arch Linux** user, then you can install ripgrep from the official repos:
```

sources/p017.md, bytes [11086, 11219):

```text
If you're a **Gentoo** user, you can install ripgrep from the
[official repo](https://packages.gentoo.org/packages/sys-apps/ripgrep):
```

sources/p017.md, bytes [11256, 11336):

```text
If you're a **Fedora** user, you can install ripgrep from official
repositories.
```

sources/p017.md, bytes [11374, 11486):

```text
If you're an **openSUSE** user, ripgrep is included in **openSUSE Tumbleweed**
and **openSUSE Leap** since 15.1.
```

sources/p017.md, bytes [11527, 11663):

```text
If you're a **RHEL/CentOS 7/8** user, you can install ripgrep from
[copr](https://copr.fedorainfracloud.org/coprs/carlwgeorge/ripgrep/):
```

sources/p017.md, bytes [11842, 11989):

```text
If you're a **Nix** user, you can install ripgrep from
[nixpkgs](https://github.com/NixOS/nixpkgs/blob/master/pkgs/tools/text/ripgrep/default.nix):
```

sources/p017.md, bytes [12086, 12306):

```text
If you're a **Debian** user (or a user of a Debian derivative like **Ubuntu**),
then ripgrep can be installed using a binary `.deb` file provided in each
[ripgrep release](https://github.com/BurntSushi/ripgrep/releases).
```

sources/p017.md, bytes [12653, 12818):

```text
If you're an **Ubuntu Cosmic (18.10)** (or newer) user, ripgrep is
[available](https://launchpad.net/ubuntu/+source/rust-ripgrep) using the same
packaging as Debian:
```

sources/p017.md, bytes [13133, 13264):

```text
If you're a **FreeBSD** user, then you can install ripgrep from the
[official ports](https://www.freshports.org/textproc/ripgrep/):
```

sources/p017.md, bytes [13297, 13421):

```text
If you're an **OpenBSD** user, then you can install ripgrep from the
[official ports](http://openports.se/textproc/ripgrep):
```

sources/p017.md, bytes [13455, 13562):

```text
If you're a **NetBSD** user, then you can install ripgrep from
[pkgsrc](http://pkgsrc.se/textproc/ripgrep):
```

sources/p017.md, bytes [13597, 13758):

```text
If you're a **Haiku x86_64** user, then you can install ripgrep from the
[official ports](https://github.com/haikuports/haikuports/tree/master/sys-apps/ripgrep):
```

sources/p017.md, bytes [13794, 13938):

```text
If you're a **Haiku x86_gcc2** user, then you can install ripgrep from the
same port as Haiku x86_64 using the x86 secondary architecture build:
```

sources/p017.md, bytes [13978, 14049):

```text
If you're a **Rust programmer**, ripgrep can be installed with `cargo`.
```

## development: removed repetition.sentence-openers (nonactionable)

The full antecedents distinguish installation variants even though their first three words repeat. No instruction can be removed merely because it shares that opening.

Diagnostic index: 90.

sources/p017.md, bytes [10276, 10380):

```text
If you're a **macOS Homebrew** or a **Linuxbrew** user, then you can install
ripgrep from homebrew-core:
```

sources/p017.md, bytes [10414, 10559):

```text
If you're a **MacPorts** user, then you can install ripgrep from the
[official ports](https://www.macports.org/ports.php?by=name&substr=ripgrep):
```

sources/p017.md, bytes [10598, 10734):

```text
If you're a **Windows Chocolatey** user, then you can install ripgrep from the
[official repo](https://chocolatey.org/packages/ripgrep):
```

sources/p017.md, bytes [10769, 10933):

```text
If you're a **Windows Scoop** user, then you can install ripgrep from the
[official bucket](https://github.com/ScoopInstaller/Main/blob/master/bucket/ripgrep.json):
```

sources/p017.md, bytes [10968, 11055):

```text
If you're an **Arch Linux** user, then you can install ripgrep from the official repos:
```

sources/p017.md, bytes [11086, 11219):

```text
If you're a **Gentoo** user, you can install ripgrep from the
[official repo](https://packages.gentoo.org/packages/sys-apps/ripgrep):
```

sources/p017.md, bytes [11256, 11336):

```text
If you're a **Fedora** user, you can install ripgrep from official
repositories.
```

sources/p017.md, bytes [11374, 11486):

```text
If you're an **openSUSE** user, ripgrep is included in **openSUSE Tumbleweed**
and **openSUSE Leap** since 15.1.
```

sources/p017.md, bytes [11527, 11663):

```text
If you're a **RHEL/CentOS 7/8** user, you can install ripgrep from
[copr](https://copr.fedorainfracloud.org/coprs/carlwgeorge/ripgrep/):
```

sources/p017.md, bytes [11842, 11989):

```text
If you're a **Nix** user, you can install ripgrep from
[nixpkgs](https://github.com/NixOS/nixpkgs/blob/master/pkgs/tools/text/ripgrep/default.nix):
```

sources/p017.md, bytes [12086, 12306):

```text
If you're a **Debian** user (or a user of a Debian derivative like **Ubuntu**),
then ripgrep can be installed using a binary `.deb` file provided in each
[ripgrep release](https://github.com/BurntSushi/ripgrep/releases).
```

sources/p017.md, bytes [12653, 12818):

```text
If you're an **Ubuntu Cosmic (18.10)** (or newer) user, ripgrep is
[available](https://launchpad.net/ubuntu/+source/rust-ripgrep) using the same
packaging as Debian:
```

sources/p017.md, bytes [13133, 13264):

```text
If you're a **FreeBSD** user, then you can install ripgrep from the
[official ports](https://www.freshports.org/textproc/ripgrep/):
```

sources/p017.md, bytes [13297, 13421):

```text
If you're an **OpenBSD** user, then you can install ripgrep from the
[official ports](http://openports.se/textproc/ripgrep):
```

sources/p017.md, bytes [13455, 13562):

```text
If you're a **NetBSD** user, then you can install ripgrep from
[pkgsrc](http://pkgsrc.se/textproc/ripgrep):
```

sources/p017.md, bytes [13597, 13758):

```text
If you're a **Haiku x86_64** user, then you can install ripgrep from the
[official ports](https://github.com/haikuports/haikuports/tree/master/sys-apps/ripgrep):
```

sources/p017.md, bytes [13794, 13938):

```text
If you're a **Haiku x86_gcc2** user, then you can install ripgrep from the
same port as Haiku x86_64 using the x86 secondary architecture build:
```

sources/p017.md, bytes [13978, 14049):

```text
If you're a **Rust programmer**, ripgrep can be installed with `cargo`.
```

## development: removed repetition.sentence-openers (nonactionable)

The grouped sentences describe different schema operations, measurement setups, or table states in separate sections. Their repeated beginnings do not establish duplicate information.

Diagnostic index: 100.

sources/p025.md, bytes [3331, 3433):

```text
Ptah does not plan that change, and the `PG307`
lint rule reports it in a migration that contains one.
```

sources/p025.md, bytes [4521, 4598):

```text
Ptah does not refresh materialized views, and a declaration cannot ask it to.
```

sources/p025.md, bytes [18186, 18408):

```text
Ptah does not
record whether an identifier was quoted, so `CREATE TABLE ORDERS` and
`CREATE TABLE "ORDERS"` reach it as the same declaration while PostgreSQL reads
them as two different relations — `orders` and `ORDERS`.
```

Diagnostic index: 105.

sources/p025.md, bytes [8999, 9304):

```text
Measured on PostgreSQL 17.10, on a database holding one table and roles nothing
grants anything to, `ptah db read` went from four `CREATE ROLE` statements to
none, and `ptah-compat schema apply --dry-run` against an empty database in a
**second** cluster went from planning three of them to planning none.
```

sources/p025.md, bytes [12795, 12919):

```text
Measured on PostgreSQL 17.10, a cluster whose superuser is `admin` accepts the
statement and the role appears in `pg_roles`.
```

sources/p025.md, bytes [33386, 33543):

```text
Measured on PostgreSQL 17 behind PgBouncer 1.25.2 in transaction mode, one
server reached two ways, the pooled answer beside the direct one for the same
URL:
```

Diagnostic index: 111.

sources/p025.md, bytes [15168, 15220):

```text
A table that is being dropped is not disabled first.
```

sources/p025.md, bytes [15221, 15392):

```text
A table that
declares a policy without declaring enablement is enabled with the table,
because `CREATE POLICY` on a table whose row-level security is off protects
nothing.
```

sources/p025.md, bytes [26311, 26451):

```text
A table that does not exist in the database yet is a
separate case and stays transactional — the migration creates it, so it starts
empty.
```

## confirmation: removed repetition.near-sentence (nonactionable)

The completion entries write scripts for bash, fish and zsh in distinct command rows. The pre-output c03-c01 control includes these actual locations; shell-specific outputs justify the repeated description.

Diagnostic index: 1.

sources/c03.md, bytes [4182, 4297):

```text
Writes a bash completion script to stdout, generated from the command tree; it opens no database and writes no file
```

sources/c03.md, bytes [4333, 4448):

```text
Writes a fish completion script to stdout, generated from the command tree; it opens no database and writes no file
```

sources/c03.md, bytes [4646, 4760):

```text
Writes a zsh completion script to stdout, generated from the command tree; it opens no database and writes no file
```
