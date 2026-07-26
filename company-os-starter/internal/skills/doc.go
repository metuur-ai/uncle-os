// Package skills merges the four skill layers.
//
// Responsibility: layer merge order, shadowing detection, and `extends`
// resolution behind `skills list`, plus validate's gate 7.
//
// It is the port of bin/company-os:751-919 and of the two call sites that gate
// uses (`:1074-1075`). The Python cluster computes structured facts and
// concatenates them into prose inside the detection loop (`:857-865`); here the
// detection loops produce Conflict records and a single pure Message function
// composes the sentence from typed fields (R-2.8, R-2.12). The counts in the
// clean line reach the text output as ints through model.Fields (R-2.3).
package skills
