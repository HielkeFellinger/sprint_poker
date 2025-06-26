
GOTH STACK

templ:

go > 1.24
go install github.com/a-h/templ/cmd/templ@latest

local
go get -tool github.com/a-h/templ/cmd/templ@latest

templ generate --watch

air