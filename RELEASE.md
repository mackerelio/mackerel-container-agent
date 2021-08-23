# Release Runbook

- Merge PRs to master
- Create release PR
  - Run action `Create Release PR`.
  - `mackerel-container-agent:<RELEASE-VERSION>-alpha` Docker image is automatically pushed.
- Check release PR
  - Check CHANGELOG.
  - Test `mackerel-container-agent:<RELEASE-VERSION>-alpha` Docker image.
- Merge release PR
- Tag `<RELEASE-VERSION>`
  - Run `git tag v$(make -s version)` and `git push --tags`.
  - `mackerel-container-agent:<RELEASE-VERSION>` Docker image is automatically pushed.
