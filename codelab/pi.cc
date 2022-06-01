// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <cassert>
#include <charconv>
#include <cmath>
#include <cstdlib>
#include <future>
#include <iostream>
#include <string_view>
#include <system_error>
#include <thread>
#include <utility>
#include <tuple>

#include <fmt/core.h>
#include <gmp.h>
#include <gmpxx.h>
#include <mpfr.h>

namespace {

constexpr unsigned int PREC_MIN = 1, PREC_MAX = 100000000;
constexpr unsigned int PREC_EXTRA_BITS = 16;

constexpr long A = 13591409, B = 545140134, C = 640320, D = 12;
constexpr long C3_OVER_24 = C * C * C / 24;

struct mpfr_real {
  mpfr_t v;

  mpfr_real() = delete;
  mpfr_real(const mpfr_real&) = delete;
  mpfr_real(const mpf_class &f) {
    mpfr_init2(v, f.get_prec());
    mpfr_set_f(v, f.get_mpf_t(), MPFR_RNDN);
  }
  ~mpfr_real() { mpfr_clear(v); }
  mpfr_ptr ptr() { return v; }
};

using bs_return_type = std::tuple<mpz_class, mpz_class, mpz_class>;

bs_return_type bs(const long a, const long b, const int depth);
mpfr_real chudnovsky_pi(const long prec, const long prec_bits);

struct bs_invoker {
  const int max_depth;
  bs_invoker() : max_depth(std::log2(std::thread::hardware_concurrency())) {}
  bs_invoker(const bs_invoker&) = delete;
  bs_invoker(int max_depth) : max_depth(max_depth) {}

  auto operator()(const long a, const long b, int depth) const {
    if (depth > max_depth) {
      std::promise<bs_return_type> promise;
      promise.set_value(bs(a, b, depth));
      return promise.get_future();
    }
    return std::async(std::launch::async, bs, a, b, depth);
  };
};

bs_return_type bs(const long a, const long b, const int depth) {
  mpz_class p, q, t;
  static bs_invoker invoker;

  if (b - a == 1) {
    if (a == 0) {
      p = 1;
      q = 1;
    } else {
      p = (6 * a - 5);
      p *= (2 * a - 1);
      p *= (6 * a - 1);
      q = a;
      q = q * a * a * C3_OVER_24;
    }
    t = p * (A + mpz_class(B) * a);
    if (a % 2)
      t = -t;
  } else {
    auto m = (a + b) / 2;
    auto f1 = invoker(a, m, depth + 1);
    auto f2 = invoker(m, b, depth + 1);
    auto [p1, q1, t1] = f1.get();
    auto [p2, q2, t2] = f2.get();
    p = p1 * p2;
    q = q1 * q2;
    t = t1 * q2 + p1 * t2;
  }
  return {p, q, t};
}

mpfr_real chudnovsky_pi(const long prec, const long prec_bits) {
  const double digits_per_term =
      std::log10(static_cast<double>(C3_OVER_24) / 6 / 2 / 6);
  const long terms = prec / digits_per_term + 2;
  fmt::print(stderr, "Number of terms = {}, digits per term = {}\n", terms,
             digits_per_term);
  auto [p, q, t] = bs(0, terms, 0);

  std::cerr << "Summation series complete. Final steps..." << std::endl;
  mpf_class q1(std::move(q), prec_bits);
  q1 = q1 * (C / D) / t;
  mpf_class sqrtC(sqrt(mpf_class(C, prec_bits)));
  return mpf_class(q1 * sqrtC);
}

void print_usage() {
  fmt::print(R"EOS(Usage: ./pi <number of digits>

Number of digits must be [{}, {}].
)EOS",
             PREC_MIN, PREC_MAX);
}

} // namespace

int main(int argc, char *argv[]) {
  unsigned int prec = 100;

  if (argc != 2) {
    print_usage();
    return 1;
  }
  const std::string_view arg(argv[1]);
  auto [ptr, ec] = std::from_chars(arg.begin(), arg.end(), prec);
  if (ptr != arg.end() && ec != std::errc()) {
    print_usage();
    return 1;
  }
  if (prec < PREC_MIN || PREC_MAX < prec) {
    print_usage();
    return 1;
  }

  const unsigned long prec_bits = prec * std::log2(10.) + PREC_EXTRA_BITS;
  fmt::print(stderr, "Calculating {} digits of pi...\n", prec);
  fmt::print(stderr, "Internal precision = {} bits\n", prec_bits);
  auto pi = chudnovsky_pi(prec, prec_bits);

  // Use MPFR to round to -inf.
  mpfr_printf("%.*RDf\n", (mpfr_prec_t)prec, pi.ptr());
  return 0;
}
