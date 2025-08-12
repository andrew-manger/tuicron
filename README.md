# tuicron

### Background
I was having trouble finding a good crontab tool (sometimes systemd is too extra). 
While it's not hard to use `crontab -e` to create and edit cron jobs, there is something a bit lackluster about defining system jobs using a text file.
I wanted a TUI that could read, edit, augment, and verify crontabs. The couple of projects I found didn't fit my needs to I decided to grab some chatbots and try taking it on myself.
I love the [charmbracelet](https://github.com/charmbracelet) libraries; [bubbletea](https://github.com/charmbracelet/bubbletea) makes beautiful TUIs. 
After some trial and error (and ultimately moving to replit), I got a working app :)

### Dependencies
[go](https://go.dev/doc/install)

### Installation
Run `go install github.com/andrew-manger/tuicron`

### Documentation
See [replit.md](https://github.com/andrew-manger/tuicron/blob/main/replit.md)
