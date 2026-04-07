#!/bin/bash

agent_demon() {
  # 将 components/collect/agent_demo/main.go 构建成 linux 二进制到 bin 目录
  mkdir -p bin
  rm -rf bin/agent_demo
  GOOS=linux GOARCH=amd64 go build -o bin/agent_demo components/collect/agent_demo/main.go

  cp bin/agent_demo /tmp/agent_demo
}

go_test() {
  local test_path="${1:-./...}"
  echo "Running tests with coverage for: $test_path"
  go test -coverprofile=coverage.out -covermode=atomic "$test_path" -timeout 60s

  echo "Coverage report:"
  go tool cover -func=coverage.out

  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
    THRESHOLD=70

    if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
        echo "[X] Coverage $COVERAGE% is below threshold $THRESHOLD%"
        set -ex
        exit 1
    else
        echo "Coverage $COVERAGE% meets threshold $THRESHOLD%"
    fi
    set -ex
}

go_vet() {
  local vet_path="${1:-./...}"
  echo "Running vet for: $vet_path"
  go vet "$vet_path" -timeout 60s
}

custom_vet() {
  go build -o custom_checker analysis/cmd/main.go
  local vet_path="${1:-./...}"
  echo "Running custom vet for: $vet_path"
  ./custom_checker "$vet_path"
}

main() {
  case "$1" in
    agent_demo)
      agent_demo
      ;;
    test)
      shift
      go_test $@
      ;;
    vet)
      shift
#      go_vet $@
      custom_vet $@
      ;;
    *)
      echo "done!"
      ;;
  esac
}

main "$@"