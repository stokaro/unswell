# GitHub settings

The settings below were applied and checked through the GitHub API on September 7,
2026. Administrators can change them; inspect the live settings during release review.

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
| Main status checks | All three Native jobs and Quality, strict | The branch must be current with passing checks |
| Main pull-request review | One approving review, dismiss stale approvals | Reviewed changes are required for normal merges |
| Force pushes and deletion of main | Disabled | Main history is protected |
| Linear history and resolved conversations | Required | Merge commits and unresolved review conversations block normal merges |
| Administrator enforcement | Disabled | The sole initial maintainer retains an explicit emergency bypass |

Repository administrators can change these settings. CI code in pull requests is
not a security boundary against a maintainer who can replace it. Workflow tokens
have read-only contents permission except the release publication job.
