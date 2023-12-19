echo "Swagger init"
swag init

echo "templ generate"
templ generate

echo "build"
go build

echo "run"
./BibleSearch