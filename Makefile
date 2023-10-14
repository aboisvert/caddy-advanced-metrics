run:
	xcaddy run --config examples/Caddyfile

build:
	#xcaddy build --with github.com/aboisvert/caddy-advanced-metrics@latest --output dist/caddy
	xcaddy build --with github.com/aboisvert/caddy-advanced-metrics@latest=. --output dist/caddy