OP_TF = op run --env-file=.env.terraform --

.PHONY: install-prereq-flox check-prereq run build test vet tidy vuln check docker-build docker-run clean terraform-init terraform-plan terraform-apply terraform-destroy

# ── Prerequisites ─────────────────────────────────────────────────────────────

install-prereq-flox:
	@which flox > /dev/null || (echo "ERROR: Flox not found. Install from https://flox.dev" && exit 1)
	@echo "Installing prerequisites via Flox..."
	flox install go terraform doctl _1password-cli gh
	@which docker > /dev/null && echo "✓ Docker already installed" || echo "⚠ Docker not found — install Docker Desktop from https://www.docker.com/products/docker-desktop"
	@echo ""
	@$(MAKE) check-prereq

check-prereq:
	@echo "Checking prerequisites..."
	@which go > /dev/null && echo "✓ Go        $(shell go version)" || echo "✗ Go        not found — run: make install-prereq-flox"
	@which op > /dev/null && echo "✓ op CLI    $(shell op --version)" || echo "✗ op CLI    not found — run: make install-prereq-flox"
	@which terraform > /dev/null && echo "✓ Terraform $(shell terraform --version | head -1)" || echo "✗ Terraform not found — run: make install-prereq-flox"
	@which doctl > /dev/null && echo "✓ doctl     $(shell doctl version | head -1)" || echo "✗ doctl     not found — run: make install-prereq-flox"
	@which gh > /dev/null && echo "✓ gh CLI    $(shell gh --version | head -1)" || echo "✗ gh CLI    not found — run: make install-prereq-flox"
	@which docker > /dev/null && echo "✓ Docker    $(shell docker --version)" || echo "✗ Docker    not found — https://www.docker.com/products/docker-desktop"
	@gh auth status > /dev/null 2>&1 && echo "✓ gh auth   authenticated" || echo "✗ gh auth   not authenticated — run: gh auth login"

# ── Development ───────────────────────────────────────────────────────────────

run:
	go run .

build:
	go build -o bin/debravinh .

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Run before every commit/push — catches vet issues, test failures, and build breaks
check: vet test build

# ── Docker ────────────────────────────────────────────────────────────────────

docker-build:
	docker build -t debravinh:local .

docker-run:
	docker run --rm -p 8080:8080 debravinh:local

clean:
	rm -rf bin
	docker rmi debravinh:local 2>/dev/null || true

# ── Terraform ─────────────────────────────────────────────────────────────────

terraform-init:
	cd terraform && terraform init

terraform-plan:
	cd terraform && $(OP_TF) terraform plan

terraform-apply:
	cd terraform && $(OP_TF) terraform apply

terraform-destroy:
	@echo "⚠ Destroying the DO app and domain — are you sure?"
	cd terraform && $(OP_TF) terraform destroy
