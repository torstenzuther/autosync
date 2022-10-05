![](logo.svg)

**Autosync** is an administration tool to automatically track changes across multiple files and / or folders and push them to a git repository.

In its current state **Autosync** is a command line tool which can be run directly and waits for EOF input (CTRL-D). On startup it syncs all files with the configured git repository and then constantly watches file changes that will be synced automatically.

## Configuration ⚙

**Autosync** is configured via its `config.json` file.

```
{
  "git-repo": {
    "url": "https://github.com/torstenzuther/autosync-test.git",
    "auth": {
      "username": "torsten.zuther@web.de",
      "password": ""
    }
  },
  "path-mappings": [
    {
      "path": "a/d",
      "pattern": "./test/*"
    }
  ],
  "processing": {
    "debounce": "1s",
    "event-channel-size": 10000
  }
}
```

### Git settings 🖧
  
Git settings include the repository URL and the optional credentials for HTTP Basic authentication.
```
{
  "git-repo": {
    "url": "https://github.com/torstenzuther/autosync-test.git",
    "auth": {
      "username": "torsten.zuther@web.de",
      "password": ""
    }
  },
  ...
```

### Path mappings 📁

Path mappings configure which local directory paths / file patterns (*pattern*) are mapped to the paths of the remote git repository (*path*).

```
...
  "path-mappings": [
    {
      "path": "a/d",
      "pattern": "./test/*"
    }
  ],
...
```
At the current implementation no sub folders are included.

### Other settings 🕰

Other settings include the debounce interval or the max commit frequency (*debounce*) and the *event-channel-size* that is a buffer size of file watcher events.

## Build 👷

You can build **autosync** yourself by running `go build` or you can download one of the pre-built releases.

## Run 🏃

Make sure that the configuration file is in the same folder as the executable and just run the application.
