package ui

var Art = map[string]string{
	"Axolotl": `
     ~^~^~
    ( o.o )~
   </ | | \>
     d   b
    ~~~~~~~`,

	"Capybara": `
      ___
    /     \
   | ^   ^ |
   |  ( )  |
    \_____/
    /|   |\`,

	"Goblin": `
      /\  /\
     ( o  o )
      > {} <
     /|/  \|\
    (_|    |_)`,

	"Pangolin": `
     ____
    /@@@@\___
   |@@@@@@   >
   |@@@@@@__/
    \@@@@/
     ~~~~`,

	"Tardigrade": `
    o _ _ o
   /       \
  | ( o o ) |
   \  __  /
    \_/\_/
   o      o`,

	"Blobfish": `
      ___
    /     \
   |  . .  |
   |   ~   |
    \_   _/
      ~~~`,

	"Mantis Shrimp": `
     \  |  /
    --O   O--
      |===|
     /|   |\
    / |   | \
      d   d`,

	"Nudibranch": `
    *  *  *
   ~~~~~~~~
  | ~    ~ |
  |  ^  ^  |
   \      /
    ~~~~~~`,

	// Rare pool
	"Salamander": `
       __
      /  \~~
     | ^^ |
      \  /
    ~~~||~~
       ()`,

	"Dire Wolf": `
      /\  /\
     /  \/  \
    | >    < |
    |  \__/  |
     \_/  \_/
     /|    |\`,

	"Thunderbird": `
     \    /
    --\  /--
      |\/|
     /|  |\
    / |  | \
    ~zZzZz~`,

	"Crystal Fox": `
      /\  /\
     *  \/  *
    | .    . |
    |   <>   |
     *______*
      | ** |`,

	"Chimera": `
    /\ ^ /\
   ( o|o|o )
    \  Y  /
    /|/ \|\
   / |   | \
   ~~~~~~~~~`,

	"Shadow Lynx": `
     . /\  /\ .
      (  ..  )
     . | {} | .
       / || \
      .  ..  .`,

	// Legendary pool
	"Phoenix": `
     * \|/ *
    -- ~~~ --
   ( > @ @ < )
    -  ^^^  -
     \|/|\|/
      *****`,

	"Leviathan": `
    ~~~~~/\
   /  O    \~~~~
  |   ~~~~   ===
   \        /
    ~~~~~~~~`,

	"Basilisk": `
      /\    /\
     /  \  /SS\
    | @  \/  @ |
    |  \~~~~/ |
     \  ====  /
      --------`,

	"Void Cat": `
    .  /\  /\  .
   . (  ??  ) .
    . | .. | .
   .   \  /   .
    .   ||   .
     . .  . .`,

	"Kraken": `
    \  ~~~~  /
     \ O  O /
      |    |
    ~/| /\ |\~
   / /|/  \|\ \
  ~ ~ ~    ~ ~ ~`,

	"Celestial Beetle": `
      *****
     *@  @*
    * /~~\ *
    *| ** |*
    * \~~/ *
     *****`,

	"Mimic": `
    .-------.
    |       |
    |  .go  |
    |       |
    | >   < |
    '---V---'`,

	"Wyrm": `
      /\
     /  \___
    | o    /
     \  __/
      \/  ___
       \_/  /
         \_/`,
}

func GetArt(species string) string {
	if art, ok := Art[species]; ok {
		return art
	}
	return `
     ???
    (o.o)
     ~~~`
}
