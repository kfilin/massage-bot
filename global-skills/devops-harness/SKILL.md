# Skill: DevOps Harness

**Description**: Expert in project infrastructure, deployment scripts, and environment isolation.

## Expertise

- **Deployment**: `deploy_home_server.sh` and `deploy_test_server.sh` lifecycle.
- **Isolation**: Managing the "Twin Strategy" (separate folders/ports for Test vs Prod).
- **Security**: SSH secrets, Docker networks (`caddy-test-net`), and DB hardening.

## Protocol

- **Safety First**: Verify environment (Prod vs Test) before running scripts.
- **Mirroring**: Enforce GitHub -> GitLab mirror flow via Rule `no-server-commits`.
- **Metrics**: Integrate health checks into all deployment flows.
