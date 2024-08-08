# Synced Drawing Canvas

After reading
[Scaling One Million Checkboxes](https://eieio.games/essays/scaling-one-million-checkboxes/)
I was inspired to create something similar. My ideas were a synced drawing
canvas or a keyboard where every sound will be propagated to all devices that
opened the page as well.

The current implementation works but could be improved in a few areas. Here are
some to dos

- use message broker like kafka or redis pub/sub
- run it on ec2
- on iPhone the selected button hasn't got round corners initially
- with high latency drawing is lagging behind. already draw line and do not wait
  for downstream message with drawing

## Usage

```sh
git clone https://github.com/PhilippReinke/synced-canvas.git
cd synced-canvas
go run .
```

Then open `localhost:8080` or `<local ip>:8080` from your mobile and the drawing
will be synced.
