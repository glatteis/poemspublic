# Hi there

You have found the backend repository for the poem printer. Exploration of this is mostly up to you, but I advise you to check out the following:

- `weather_config.toml` is the file for OpenWeatherMap API configuration. It should be documented okay.
- `data/` is the folder where you put your .json poem files. A sample is in this folder.
- `crawler/` contains a python crawler that crawls poems from wikisource and puts them into the `data/` folder. Feel free to throw it away if you don't need it.
- The rest is just a go program. Get the libraries with `go get`, build it with `go build`, and run it by executing the resulting binary (it takes a port argument: `-port <port>`).
It needs `wkhtmltoimage` (from `wkhtmltopdf`) installed. It also supports it being installed via `xvfb-run`. Please look at `imggenerator/generator.go`.
- For testing the poem endpoint, visit `/name` and then `/poem?name=<name>`.
- For testing weather endpoint, visit `/weather`.
- Templates are in German. Modify them if you want them in another language :)