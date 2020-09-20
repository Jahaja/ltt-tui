# ltt-tui
Terminal-UI for ltt: https://github.com/Jahaja/ltt

# Usage

`go build`

`./ttl-tui -uri http://uri-to-ltt-server.com/`

Navigate using the vi/vim keys (`h`, `j`, `k`, `l`) or the arrow keys.
The `q` key exits the program.

### Buttons
* The `Set` button will start or modify the number of users to simulate
* The `Stop` button will set the user number to 0 and this stop the test.
* The `Reset` button will reset all the statistics.
* The `Errors/Tasks` button toggles the table between tasks or errors.

The Tasks table is sortable by navigating to a column and pressing Enter.