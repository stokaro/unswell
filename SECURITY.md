# Security

Report vulnerabilities privately through GitHub's **Report a vulnerability** action
on this repository's Security tab. Do not include private source text, access tokens,
or confidential reports in public issues.

Only the latest alpha is supported. Alpha releases are experimental and have not
received an independent security audit. Reports can contain source text when
`--include-source` is selected; protect them as you would the original documents.

The runtime works offline. Custom Go rules and NLP providers execute trusted code
inside the calling process. They are not a sandbox for untrusted extensions.
