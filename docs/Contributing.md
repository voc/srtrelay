# Contributing
You can use the local nix flake as a development environment
```
nix develop
```

Please run tests before creating a pull request
```
make lint test
```

## General concepts
  - srtrelay should just be a 1:n multiplexer, one publisher (push) to multiple subscribers (pull)
  - Don't try to reimplement functionality already present elsewhere in the stack (e.g. remuxing/transcoding)
  - Allow any data to be relayed, not just MPEG-TS
