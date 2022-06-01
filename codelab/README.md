# How to build

Install prerequisites.
This program uses [GMP](https://gmplib.org/), [MPFR](https://www.mpfr.org/), and [{fmt}](https://github.com/fmtlib/fmt).
You also need a working C++17 compiler (tested with GCC 10 on Debian 11).

On Debian, run:

```bash
sudo apt update
sudo apt install -y build-essential libgmp-dev libmpfr-dev libfmt-dev
```

Now compile the source:
```bash
c++ -opi pi.cc -std=c++17 -O3 -march=native -lgmp -lmpfr -lpthread -lfmt
```

You'll have `pi` executable file in the current directory.

# How to run

The program takes one argument, number of decimal digits to calculate. For example, to calculate 100 digits of pi, run:

```bash
./pi 100
```
