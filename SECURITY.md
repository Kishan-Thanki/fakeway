# Security Policy

## Reporting a vulnerability

Do not report suspected vulnerabilities in public issues. Use the private vulnerability reporting feature for this repository if enabled. Otherwise, contact the maintainer privately through the GitHub profile. Include the affected version or commit, reproduction steps, and potential impact. Do not include real credentials or production data.

## Supported versions

The latest released version and the main branch are the primary targets for security fixes. Older releases may not receive backported fixes. Upgrade to the latest release where possible.

## Trust boundary

Fakeway is intended for local development, testing, and API mocking. It is not intended to be exposed to untrusted networks without appropriate safeguards.

Only run Fakeway in environments where you understand who can reach the server and what requests it accepts. When binding to non-loopback interfaces, restrict network access to trusted clients.

Do not assume that a mock server is safe to expose publicly simply because it does not connect to a production backend.

## Request handling and responses

Fakeway may receive arbitrary HTTP requests and return configured mock responses. Treat request data as untrusted input. Avoid using real credentials, production data, or sensitive information in mock configurations and responses.

If Fakeway supports request matching, templating, or dynamic response generation, review those configurations carefully before exposing the server to untrusted clients.

## Network exposure

The security impact of running Fakeway depends on its listener configuration, request-handling behavior, and deployment environment. Use loopback binding for local-only testing where supported.

When running Fakeway in a container, CI environment, or shared network, restrict exposed ports and apply appropriate network controls. Do not assume that containerization alone provides sufficient isolation.

## Reporting security-sensitive behavior

Please report unexpected access to sensitive data, unintended network exposure, denial-of-service conditions, or other security-relevant behavior through the private reporting process described above.
