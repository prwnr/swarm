# Swarm - a stream monitor

Redis Streams monitor package written in Go.

It is meant to serve as a handy streams debugging tool, providing interactive terminal
interface to listen on all active streams and read their messages.

In addition to streams monitoring it provides automatic listening for [Laravel Streamer](https://github.com/prwnr/laravel-streamer) package.

Navigation: 
1) `1` and `2` between tabs (if listening is active)
2) `up` and `down` arrows to walk over rows
3) `enter` to select row 
4) `escape` to get back to left column when stream was selected before

For Streamer messages copying on Linux install `xsel` command.

**!In Works!** 
