#!/usr/bin/env bash
set -euo pipefail

base_ref="${DOCKER_BUILD_BASE:-origin/main}"
if ! git rev-parse --verify "${base_ref}^{commit}" >/dev/null 2>&1; then
  echo "docker-build: cannot resolve base ref ${base_ref}" >&2
  exit 1
fi

declare -A dockerfiles=()
declare -A affected=()

register_production_dockerfile() {
  local path="$1"
  case "$path" in
    deploy/docker/api.Dockerfile) dockerfiles[api]="$path" ;;
    deploy/docker/worker.Dockerfile) dockerfiles[worker]="$path" ;;
    deploy/docker/jobs.Dockerfile) dockerfiles[jobs]="$path" ;;
    deploy/docker/dbprovision.Dockerfile) dockerfiles[dbprovision]="$path" ;;
    frontend/apps/web/Dockerfile) dockerfiles[web]="$path" ;;
    apps/docx-renderer/Dockerfile) dockerfiles[docx]="$path" ;;
    vendor/*) ;;
    *)
      echo "docker-build: unclassified tracked Dockerfile: ${path}" >&2
      exit 1
      ;;
  esac
}

while IFS= read -r tracked_path; do
  case "$tracked_path" in
    */Dockerfile|*.Dockerfile|*.dockerfile)
      register_production_dockerfile "$tracked_path"
      ;;
  esac
done < <(git ls-files)

mark() {
  affected["$1"]=1
}

mark_go_images() {
  mark api
  mark worker
  mark jobs
  mark dbprovision
}

mark_node_images() {
  mark web
  mark docx
}

mark_all_images() {
  mark_go_images
  mark_node_images
}

while IFS= read -r changed_path; do
  case "$changed_path" in
    deploy/docker/api.Dockerfile|apps/api/*|api/openapi/*) mark api ;;
    deploy/docker/worker.Dockerfile|apps/worker/*) mark worker ;;
    deploy/docker/jobs.Dockerfile|apps/jobs/*) mark jobs ;;
    deploy/docker/dbprovision.Dockerfile|apps/dbprovision/*) mark dbprovision ;;
    frontend/apps/web/*) mark web ;;
    apps/docx-renderer/*) mark docx ;;
    internal/*|db/*|go.mod|go.sum) mark_go_images ;;
    .dockerignore|scripts/check-docker-build.sh) mark_all_images ;;
    packages/*|package.json|pnpm-lock.yaml|pnpm-workspace.yaml|.nvmrc) mark_node_images ;;
  esac
done < <(git diff --name-only --diff-filter=ACDMRTUXB "${base_ref}...HEAD")

if [[ ${#affected[@]} -eq 0 ]]; then
  echo "docker-build: no affected production image"
  exit 0
fi

tag="${GITHUB_SHA:-local}"
for artifact in api worker jobs dbprovision web docx; do
  if [[ -z "${affected[$artifact]+set}" ]]; then
    continue
  fi
  dockerfile="${dockerfiles[$artifact]-}"
  if [[ -z "$dockerfile" ]]; then
    echo "docker-build: affected artifact ${artifact} has no tracked production Dockerfile" >&2
    exit 1
  fi
  echo "docker-build: building ${artifact} from ${dockerfile}"
  docker build --file "$dockerfile" --tag "metaldocs-ci-${artifact}:${tag}" .
done
