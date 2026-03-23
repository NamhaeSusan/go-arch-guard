Summary:
- Updated README orchestration docs to match current `CheckDomainIsolation` behavior.
- Clarified that orchestration must use domain root imports for domains, but may import other non-domain internal packages.
- Clarified that `CheckStructure` is responsible for whether extra internal packages are allowed in the tree.

Files changed:
- README.md

Verification:
- Reviewed `git diff -- README.md`
- Re-read updated README orchestration section and import matrix
