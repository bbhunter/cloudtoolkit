package runner

import "fmt"

const banner = `
  ▄▄▄ ▄▄▄▄▄▄▄ ▄    ▄
▄▀   ▀   █    █  ▄▀
█        █    █▄█
█        █    █  █▄
 ▀▄▄▄▀   █    █   ▀▄
                      v%s
`

// Version is the current version of cloudtoolkit
const Version = `0.1.6`

// showBanner is used to show the banner to the user
func ShowBanner() {
	fmt.Printf(banner, Version)
}
