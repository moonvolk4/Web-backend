Redis via Docker (macOS)

- Option A: Use docker-compose in this repo
  - A service named `redis` is defined; start stack and Redis will be available at 127.0.0.1:6379 with append-only persistence.
- Option B: Run a standalone container
  - docker run -d --name redis -p 6379:6379 redis:7-alpine
  - For persistence: add -v redis_data:/data and --appendonly yes
- Optional security
  - Set a password with --requirepass <pass> and update config RedisPassword accordingly.

App configuration keys (config/config.toml)
- RedisAddr = "127.0.0.1:6379"
- RedisPassword = ""
- JWTSecret = "dev-secret-change-me"
- JWTTTLMinutes = 60
- CookieName = "access_token"

In production
- Use a strong JWTSecret and HTTPS so cookies are secure.
- Consider setting cookie Secure and SameSite settings at the reverse proxy.
