# GitHub settings

The settings below were checked through the GitHub API on September 7, 2026.
Administrators can change them; inspect the live settings during release review.

| Setting | Configured value | Effect |
| --- | --- | --- |
| Visibility | Public, already created | Source is publicly readable |
| Merge methods | Squash only | Pull requests produce one main-branch commit |
| Delete merged branches | Enabled | GitHub deletes merged pull-request head branches |
| Issues | Enabled | Public bug and feature tracking |
| Wiki and projects | Disabled | Documentation and work tracking stay in the repository |
| Secret scanning and push protection | Enabled | GitHub detects supported secret patterns and can reject affected pushes |
| Dependency alerts and security fixes | Enabled | GitHub reports dependency advisories and can create remediation PRs |
| Private vulnerability reporting | Enabled | Researchers can report vulnerabilities privately |
| Main ruleset | Active `main` ruleset, ID 22441310 | Applies to the default branch |
| Main status checks | No required checks in the current ruleset | Release jobs still depend on all six CI checks |
| Main pull requests | Required, squash only, zero required approving reviews | Changes must go through a pull request |
| Force pushes and deletion of main | Disabled | Main history is protected |
| Linear history | Required | Merge commits cannot enter main |
| Ruleset bypass actors | None | The current ruleset provides no bypass actor |

The original classic branch protection was replaced during alpha preparation.
The current ruleset does not require approval, stale-review dismissal, or resolved
review conversations. Passing CI remains part of the project's release acceptance;
the ruleset itself does not enforce those checks on a pull request.

The separate Homebrew tap currently requires four native installation checks and
one approving review. The Action repository currently requires three native test
checks and one approving review. Both protect main from force pushes and deletion.

The distribution automation needs auto-merge enabled in both repositories and a
review exception for the `ptah-publish` app alone. Other authors retain the review
requirement. The Action's three published-release consumer checks must also become
required. These configuration changes are pending explicit authorization; the
workflow implementation does not establish that the settings are already active.
The org secret and variable currently select only Ptah, and the installed app
still needs Pull requests write access. See [publishing setup](installation.md#publishing-app-setup).

Repository administrators can change these settings. CI code in pull requests is
not a security boundary against a maintainer who can replace it. Workflow tokens
have read-only contents permission by default. Release jobs receive the specific
contents, packages, or OIDC permissions needed for their publication step.
