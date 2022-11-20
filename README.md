# go-bcs

Binary Canonical Serialization (bcs) for Golang.

[![Go Reference](https://pkg.go.dev/badge/github.com/fardream/go-bcs.svg)](https://pkg.go.dev/github.com/fardream/go-bcs)

Binary Canonical Serialization (bcs) or Libra Canonical Serialization (lcs) is developed
for the shuttlered [libra/diem](https://www.diem.com/) blockchain project.

Its target is mainly rust-lang struct, although many [move-lang](https://github.com/move-language/move) based
blockchains use it as serialization format.

Given its root in rust, bcs include many structs are unavailable in golang (or move-lang), such as enum, option. See [go package website](https://pkg.go.dev/github.com/fardream/go-bcs) for more details.
