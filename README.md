A notification plugin for (mostly) Icinga 2 which manages problems as tickets in a Zammad instance


> [!WARNING]
> This is still in an experimental state and heavy developement. Not functional yet, untested and not ready for production


## Building

### Necessary tools

 * the [`golang` toolchain](https://go.dev/)

### Compiling

```
go build
```
executed in the main folder of this repository will generate an executable. If you are not on a linux system,
it should probably look like this:

```
GOOS=linux go build
```
