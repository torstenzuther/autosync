![](logo.svg)

**Autosync** is an administration tool to automatically track changes across multiple files and / or folders and push them to a git repository.

## Configuration ⚙

### File patterns 📁

* patterns
* mention that no sub folders are included

### Git settings 🖧
    
* repository URL
* init remote repository on start
* authentication settings (HTTP Basic)

### Other settings 🕰
* event debounce interval
* push debounce interval

## Build 👷

You can build **autosync** yourself by running `go build` or you can download one of the pre-built releases.

## Run 🏃

Make sure that the configuration file is in the same folder as the executable and just run the application.
You can also run it as a service (Windows / Unix).