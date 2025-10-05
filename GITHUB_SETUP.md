# GitHub Setup Guide

Quick guide to push this repository to GitHub and configure branch protection.

## Initial Push to GitHub

```bash
# Create a new repository on GitHub (web interface)
# Then push your local repository:

git remote add origin https://github.com/yourusername/blandmockapi.git
git push -u origin main
git push -u origin dev
```

## Branch Protection Setup

### Protect Main Branch

1. Go to repository Settings > Branches
2. Add branch protection rule for `main`:
   - ✓ Require a pull request before merging
   - ✓ Require approvals (at least 1)
   - ✓ Require status checks to pass before merging
     - Select: `Unit & Integration Tests`
     - Select: `Build Binaries`
     - Select: `Build Docker Images`
   - ✓ Require branches to be up to date before merging
   - ✓ Include administrators (recommended)

### Protect Dev Branch

1. Add branch protection rule for `dev`:
   - ✓ Require a pull request before merging
   - ✓ Require status checks to pass before merging
     - Select: `Unit & Integration Tests`
   - ✓ Require branches to be up to date before merging

## GitHub Actions Secrets (Optional)

If deploying to AWS from GitHub Actions:

1. Go to Settings > Secrets and variables > Actions
2. Add repository secrets:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION` (or use default in workflow)
   - `AWS_ACCOUNT_ID`

## TeamCity Integration

### VCS Root Setup

1. In TeamCity, create new VCS Root:
   - Type: Git
   - Fetch URL: `https://github.com/yourusername/blandmockapi.git`
   - Default branch: `refs/heads/main`
   - Branch specification: `+:refs/heads/(main|dev)`

2. Authentication:
   - Use GitHub personal access token or SSH key

### Build Triggers

The `.teamcity/settings.kts` is pre-configured with:
- Triggers on `main` and `dev` branches
- Build dependencies for deployment pipelines
- Automated tests before deployments

## Workflow Examples

### Feature Development
```bash
# Start from dev branch
git checkout dev
git pull origin dev

# Create feature branch
git checkout -b feature/add-webhook-support

# Make changes, commit
git add .
git commit -m "Add webhook endpoint support"

# Push and create PR to dev
git push origin feature/add-webhook-support
# Create PR on GitHub: feature/add-webhook-support → dev
```

### Release to Production
```bash
# After dev is tested and ready
git checkout main
git pull origin main

# Merge dev into main
git merge dev

# Tag the release (optional)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push to production
git push origin main
git push origin --tags
```

## Repository Settings Recommendations

### General
- ✓ Allow squash merging (recommended)
- ✗ Allow merge commits (optional)
- ✗ Allow rebase merging (optional)
- ✓ Automatically delete head branches

### Pull Requests
- ✓ Require linear history (if using squash merge)
- ✓ Require successful deployments before merging (for main)

### Collaborators
- Add team members with appropriate permissions:
  - Admin: Full access
  - Maintain: Manage settings without destructive actions
  - Write: Push to dev, create PRs
  - Read: View code only

## CI/CD Pipeline Flow

```
Feature Branch
     │
     ├─ Push
     │
     ├─ GitHub Actions: Run Tests
     │
     ├─ Create PR to dev
     │
     ▼
   Dev Branch
     │
     ├─ Merge PR
     │
     ├─ TeamCity: Run Tests + Deploy to Staging
     │
     ├─ Manual Testing
     │
     ├─ Create PR to main
     │
     ▼
  Main Branch
     │
     ├─ Merge PR (with approval)
     │
     ├─ GitHub Actions: Run All Tests + Build
     │
     ├─ TeamCity: Run Tests + Deploy to Production
     │
     ▼
  Production
```

## Monitoring & Notifications

### GitHub
- Settings > Notifications: Configure email/Slack for PR reviews
- Enable Dependabot alerts for security updates

### TeamCity
- Configure build failure notifications
- Set up Slack/email integration for deployment status

## Quick Commands Reference

```bash
# Clone repository
git clone https://github.com/yourusername/blandmockapi.git
cd blandmockapi

# Setup for development
git checkout dev
go mod download
make test

# Create feature branch
git checkout -b feature/my-feature

# Run tests locally
./scripts/test.sh

# Push feature
git push origin feature/my-feature

# After PR merged to dev, update local dev
git checkout dev
git pull origin dev

# Release to production (maintainers only)
git checkout main
git merge dev
git push origin main
```

## Troubleshooting

**Issue**: PR fails status checks
- Check GitHub Actions logs
- Run tests locally: `./scripts/test.sh`
- Ensure coverage doesn't decrease

**Issue**: Can't push to main
- Main is protected - create PR instead
- Get required approvals
- Ensure all checks pass

**Issue**: TeamCity not triggering
- Verify VCS Root configuration
- Check branch specification: `+:refs/heads/(main|dev)`
- Verify webhook configuration

## Additional Resources

- [GitHub Flow Guide](https://guides.github.com/introduction/flow/)
- [Branch Protection Rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [TeamCity VCS Roots](https://www.jetbrains.com/help/teamcity/vcs-root.html)
