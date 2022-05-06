# GoAT: Go ASCII Tool
<!--
  NOTE to maintainers
  ---
    SVG examples/ regeneration.
       go test -run . -v -write

    Github home page README.md, specific to $USER:
       sed "s,{{\.Root}},https://cdn.rawgit.com/${USER}/goat/main," README.md.tmpl >README.md

    Local preview of home page:
       sed "s,https://cdn.rawgit.com/blampe/goat/main,.," README.md.tmpl >README.md
       # See https://github.github.com/gfm/#introduction
       (echo '<!DOCTYPE html>'; marked -gfm README.md) >README.html

    The @media query from SVG may be verified in Firefox by switching between Themes
    "Light" and "Dark" in Firefox's "Add-ons Manager".
 -->

This is a Go implementation of [markdeep.mini.js]'s ASCII diagram
generation.

## Update (2022-02-07)

I hacked together GoAT a number of years ago while trying to embed some
diagrams in a Hugo project I was playing with. Through an odd twist of fate
GoAT eventually made its way into the upstream Hugo project, and if you're
using [v0.93.0] you can embed these diagrams natively. Neat!

My original implementation was certainly buggy and not on par with markdeep.
I'm grateful for the folks who've helped smooth out the rough edges, and I've
updated this project to reflect the good changes made in the Hugo fork,
including a long-overdue `go.mod`.

There's a lot I would like to do with this project that I will never get to, so
instead I recommend you look at these forks:

* [@bep] is the fork currently used by Hugo, which I expect to be more active
  over time.
* [@dmacvicar] has improved SVG/PNG/PDF rendering.
* [@sw46] has implemented a really wonderful hand-drawn style worth checking
  out.

## Usage

```bash
$ go get github.com/blampe/goat
$ cat my-cool-diagram.txt | goat > my-cool-diagram.svg
```

By default, the program reads from stdin, unless `-i infile` is given.

By default, the program writes to stdout, unless `-o outfile` is given or a
binary format with `-f` is selected.

By default, it writes in [SVG] format, unless another format is specified with
`-f`.

## TODO

- Dashed lines signaled by `:` or `=`.
- Bold lines signaled by ???.

## Examples

Here are some SVGs and the UTF-8 input they were generated from:

### Trees

![Trees Example](https://cdn.rawgit.com/blampe/goat/main/examples/trees.svg)

```
          .               .                .               .--- 1          .-- 1     / 1
         / \              |                |           .---+            .-+         +
        /   \         .---+---.         .--+--.        |   '--- 2      |   '-- 2   / \ 2
       +     +        |       |        |       |    ---+            ---+          +
      / \   / \     .-+-.   .-+-.     .+.     .+.      |   .--- 3      |   .-- 3   \ / 3
     /   \ /   \    |   |   |   |    |   |   |   |     '---+            '-+         +
     1   2 3   4    1   2   3   4    1   2   3   4         '--- 4          '-- 4     \ 4
```

### Overlaps

![Overlaps Example](https://cdn.rawgit.com/blampe/goat/main/examples/overlaps.svg)

```
           .-.           .-.           .-.           .-.           .-.           .-.
          |   |         |   |         |   |         |   |         |   |         |   |
       .---------.   .--+---+--.   .--+---+--.   .--|   |--.   .--+   +--.   .------|--.
      |           | |           | |   |   |   | |   |   |   | |           | |   |   |   |
       '---------'   '--+---+--'   '--+---+--'   '--|   |--'   '--+   +--'   '--|------'
          |   |         |   |         |   |         |   |         |   |         |   |
           '-'           '-'           '-'           '-'           '-'           '-'
```

### Line Decorations

![Line Decorations Example](https://cdn.rawgit.com/blampe/goat/main/examples/line-decorations.svg)

```
                ________                            o        *          *   .--------------.
   *---+--.    |        |     o   o      |         ^          \        /   |  .----------.  |
       |   |    '--*   -+-    |   |      v        /            \      /    | |  <------.  | |
       |    '----->       .---(---'  --->*<---   /      .+->*<--o----'     | |          | | |
   <--'  ^  ^             |   |                 |      | |  ^    \         |  '--------'  | |
          \/        *-----'   o     |<----->|   '-----'  |__|     v         '------------'  |
          /\                                                               *---------------'
```

### Line Ends

![Line Ends Example](https://cdn.rawgit.com/blampe/goat/main/examples/line-ends.svg)

```
   o--o    *--o     /  /   *  o  o o o o   * * * *   o o o o   * * * *      o o o o   * * * *
   o--*    *--*    v  v   ^  ^   | | | |   | | | |    \ \ \ \   \ \ \ \    / / / /   / / / /
   o-->    *-->   *  o   /  /    o * v '   o * v '     o * v \   o * v \  o * v /   o * v /
   o---    *---
                                 ^ ^ ^ ^   . . . .   ^ ^ ^ ^   \ \ \ \      ^ ^ ^ ^   / / / /
   |  |   *  o  \  \   *  o      | | | |   | | | |    \ \ \ \   \ \ \ \    / / / /   / / / /
   v  v   ^  ^   v  v   ^  ^     o * v '   o * v '     o * v \   o * v \  o * v /   o * v /
   *  o   |  |    *  o   \  \

   <--o   <--*   <-->   <---      ---o   ---*   --->   ----      *<--   o<--   -->o   -->*
```

### Dot Grids

![Dot Grids Example](https://cdn.rawgit.com/blampe/goat/main/examples/dot-grids.svg)

```
  o o o o o  * * * * *  * * o o *    o o o      * * *      o o o     · * · · ·     · · ·
  o o o o o  * * * * *  o o o o *   o o o o    * * * *    * o * *    · * * · ·    · · · ·
  o o o o o  * * * * *  o * o o o  o o o o o  * * * * *  o o o o o   · o · · o   · · * * ·
  o o o o o  * * * * *  o * o o o   o o o o    * * * *    o * o o    · · · · o    · · * ·
  o o o o o  * * * * *  * * * * o    o o o      * * *      o * o     · · · · ·     · · *
```
Note that '·' above is not ASCII, but rather Unicode, the MIDDLE DOT character, encoded with UTF-8.

### Large Nodes

![Large Node Example](https://cdn.rawgit.com/blampe/goat/main/examples/large-nodes.svg)

```
   .---.       .-.        .-.       .-.                                       .-.
   | A +----->| 1 +<---->| 2 |<----+ 4 +------------------.                  | 8 |
   '---'       '-'        '+'       '-'                    |                  '-'
                           |         ^                     |                   ^
                           v         |                     v                   |
                          .-.      .-+-.        .-.      .-+-.      .-.       .+.       .---.
                         | 3 +---->| B |<----->| 5 +---->| C +---->| 6 +---->| 7 |<---->| D |
                          '-'      '---'        '-'      '---'      '-'       '-'       '---'
```

### Small Grids

![Small Grids Example](https://cdn.rawgit.com/blampe/goat/main/examples/small-grids.svg)

```
       ___     ___      .---+---+---+---+---.     .---+---+---+---.  .---.   .---.
   ___/   \___/   \     |   |   |   |   |   |    / \ / \ / \ / \ /   |   +---+   |
  /   \___/   \___/     +---+---+---+---+---+   +---+---+---+---+    +---+   +---+
  \___/ b \___/   \     |   |   | b |   |   |    \ / \a/ \b/ \ / \   |   +---+   |
  / a \___/   \___/     +---+---+---+---+---+     +---+---+---+---+  +---+ b +---+
  \___/   \___/   \     |   | a |   |   |   |    / \ / \ / \ / \ /   | a +---+   |
      \___/   \___/     '---+---+---+---+---'   '---+---+---+---'    '---'   '---'
```

### Big Grids

![Big Grids Example](https://cdn.rawgit.com/blampe/goat/main/examples/big-grids.svg)

```
    .----.        .----.
   /      \      /      \            .-----+-----+-----.
  +        +----+        +----.      |     |     |     |          .-----+-----+-----+-----+
   \      /      \      /      \     |     |     |     |         /     /     /     /     /
    +----+   B    +----+        +    +-----+-----+-----+        +-----+-----+-----+-----+
   /      \      /      \      /     |     |     |     |       /     /     /     /     /
  +   A    +----+        +----+      |     |  B  |     |      +-----+-----+-----+-----+
   \      /      \      /      \     +-----+-----+-----+     /     /  A  /  B  /     /
    '----+        +----+        +    |     |     |     |    +-----+-----+-----+-----+
          \      /      \      /     |  A  |     |     |   /     /     /     /     /
           '----'        '----'      '-----+-----+-----'  '-----+-----+-----+-----+
```

### Complicated

![Complicated Example](https://cdn.rawgit.com/blampe/goat/main/examples/complicated.svg)

```
+-------------------+                           ^                      .---.
|    A Box          |__.--.__    __.-->         |      .-.             |   |
|                   |        '--'               v     | * |<---        |   |
+-------------------+                                  '-'             |   |
                       Round                                       *---(-. |
  .-----------------.  .-------.    .----------.         .-------.     | | |
 |   Mixed Rounded  | |         |  / Diagonals  \        |   |   |     | | |
 | & Square Corners |  '--. .--'  /              \       |---+---|     '-)-'       .--------.
 '--+------------+-'  .--. |     '-------+--------'      |   |   |       |        / Search /
    |            |   |    | '---.        |               '-------'       |       '-+------'
    |<---------->|   |    |      |       v                Interior                 |     ^
    '           <---'      '----'   .-----------.              ---.     .---       v     |
 .------------------.  Diag line    | .-------. +---.              \   /           .     |
 |   if (a > b)     +---.      .--->| |       | |    | Curved line  \ /           / \    |
 |   obj->fcn()     |    \    /     | '-------' |<--'                +           /   \   |
 '------------------'     '--'      '--+--------'      .--. .--.     |  .-.     +Done?+-'
    .---+-----.                        |   ^           |\ | | /|  .--+ |   |     \   /
    |   |     | Join        \|/        |   | Curved    | \| |/ | |    \    |      \ /
    |   |     +---->  o    --o--        '-'  Vertical  '--' '--'  '--  '--'        +  .---.
 <--+---+-----'       |     /|\                                                    |  | 3 |
                      v                             not:line    'quotes'        .-'   '---'
  .-.             .---+--------.            /            A || B   *bold*       |        ^
 |   |           |   Not a dot  |      <---+---<--    A dash--is not a line    v        |
  '-'             '---------+--'          /           Nor/is this.            ---
```

More examples are available [here](examples).

[@bep]: https://github.com/bep/goat/
[@dmacvicar]: https://github.com/dmacvicar/goat
[@sw46]: https://github.com/sw46/goat/
[SVG]: https://en.wikipedia.org/wiki/Scalable_Vector_Graphics
[markdeep.mini.js]: http://casual-effects.com/markdeep/
[v0.93.0]: https://github.com/gohugoio/hugo/releases/tag/v0.93.0
