# orclib

orclib is a set of libraries designed for use with Orc:
    github.com/steinarvk/orc
Together with the core Orc code, they are an opinionated
framework that attempts to make it easy to build server
binaries in Go.

## Goal

The goal of Orc is to make it possible to write very
short Go binaries -- especially server binaries -- that
come with an ever-expanding set of batteries included,
making them pleasant to debug and work with.

The Orc libraries handle all the boilerplate and
essentially initialize themselves, letting the application
code focus on the actual application logic.

## State

The Orc framework is highly experimental at the moment.
It's currently not recommended for general use.

## Libraries

The libraries (in lib/) are normal Go packages for
code that happened to be useful supporting code for
Orclib. Some of them might be generally useful.

## Modules

The modules (in module/) are special Go packages that
use the Orc framework to self-initialize. Since they
are part of a highly opinionated framework, it's not
a good idea to try to reuse them if you're not using
the Orc framework.

Feel free to fork the code if you see anything useful.

## Commands 

The commands (in command/) are reusable Cobra commands
that are generally directly applicable to managing
Orc projects and servers.

## Authorship

Orc was written by me, Steinar V. Kaldager.

This is a hobby project written primarily for my own usage.
Don't expect support for it. It was developed in my spare
time and is not affiliated with any employer of mine.

It is released as open source under the MIT license. Feel
free to make use of it however you wish under the terms
of that license.

## License

This project is licensed under the MIT License - see the
LICENSE file for details.
