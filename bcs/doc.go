// bcs (binary canonical serialization) or lcs (libra canonical serialization) is developed
// from the shuttered [libra/diem] block chain project.
//
// [bcs] defines how a struct in rust-lang can be serialized into bytes, and supports features that are
// unavaibable in golang, such as tagged union or enum.
//
// By "canonical", it means the serialization is deterministic and unique.
// On many [move-lang] based blockchains, bcs is the serialization scheme for the struct in move.
//
// See [Marshal] and [Unmarshal] for details.
//
// [libra/diem]: https://www.diem.com/
// [move-lang]: https://github.com/move-language/move
// [bcs]: https://github.com/diem/bcs
package bcs
