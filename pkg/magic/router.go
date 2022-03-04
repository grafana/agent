package magic

import (
	_ "embed"
)

//go:embed assets/ninja.js
var ninjajs []byte

//go:embed assets/ninja-footer.js
var ninjafooter []byte

//go:embed assets/ninja-action.js
var ninjaaction []byte

//go:embed assets/ninja-header.js
var ninjaheader []byte

//go:embed assets/base-styles.js
var basestyles []byte
