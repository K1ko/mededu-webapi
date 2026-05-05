param (
    $command
)

if (-not $command) {
    $command = "start"
}

$ProjectRoot = "${PSScriptRoot}/.."

$env:MEDEDU_API_ENVIRONMENT = "Development"
$env:MEDEDU_API_PORT = "8080"
$env:MEDEDU_API_MONGODB_USERNAME = "root"
$env:MEDEDU_API_MONGODB_PASSWORD = "neUhaDnes"
$env:MEDEDU_API_MONGODB_PORT = "27018"
$env:MEDEDU_API_MONGO_EXPRESS_PORT = "8082"

function mongo {
    docker compose --env-file ${ProjectRoot}/deployments/docker-compose/.env --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/mededu-api-service
        } finally {
            mongo down
        }
    }
    "openapi" {
        docker run --rm -ti -v ${ProjectRoot}:/local openapitools/openapi-generator-cli generate -c /local/scripts/generator-cfg.yaml
    }
    "mongo" {
        mongo up
    }
    default {
        throw "Unknown command: $command"
    }
}
