# Business scenarios

Godog runs every `.feature` file here against each driver (memory, jsonfile, http, desktop and mobile) with `task accept`. Tags limit where a scenario runs: `@desktop` and `@mobile` scenarios run only on their own driver, `@workspaces` ones only on memory and jsonfile, and `@plans` ones only on memory, jsonfile and http, where the tests set each workspace's plan (see `drivers` in `test/acceptance/drivers.go`). `/feature` adds a file per capability, written in the domain's language from the glossary in `SPEC.md`.
