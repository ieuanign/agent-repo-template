#!/bin/sh
# Apple-native builds vs. devbox's nix C toolchain.
#
# On macOS devbox puts a nix C toolchain into every shell in this repo: CC, CXX,
# LD, AR, NM..., NIX_CFLAGS_COMPILE, NIX_LDFLAGS, SDKROOT, DEVELOPER_DIR, plus
# nix's clang and a 2019 xcbuild `xcrun` on PATH. Xcode inherits all of it, and
# xcodebuild also reads environment variables as build settings. Two ways that
# breaks `expo run:ios`:
#
#   - NIX_CFLAGS_COMPILE puts nix's libc++ ahead of Apple's headers, so the
#     iPhoneSimulator SDK fails to compile (`unknown type name 'uint8_t'`).
#   - LD=ld makes xcodebuild link by invoking ld directly with clang-driver
#     flags, so -rpath gets no path (`-objc_abi_version '-Xlinker'`).
#
# Sourced (the devbox init_hook) this cleans the current shell; run with a
# command (`scripts/xcode-env.sh npx expo run:ios`) it cleans and execs that
# command. Only the compiler toolchain is removed — nix's node stays on PATH.

if [ "$(uname -s)" = "Darwin" ]; then
  unset CC CXX LD AR AS NM RANLIB STRIP LIBTOOL OBJCOPY OBJDUMP SIZE STRINGS \
        READELF CFLAGS CXXFLAGS CPPFLAGS LDFLAGS LD_DYLD_PATH \
        SDKROOT DEVELOPER_DIR MACOSX_DEPLOYMENT_TARGET

  # NIX_SSL_CERT_FILE is kept: it is certificate config, not toolchain config.
  for _xcode_env_var in $(env | sed -n 's/^\(NIX_[A-Za-z0-9_]*\)=.*/\1/p'); do
    [ "$_xcode_env_var" = "NIX_SSL_CERT_FILE" ] || unset "$_xcode_env_var"
  done
  unset _xcode_env_var

  PATH="$(printf '%s' "$PATH" | tr ':' '\n' \
    | grep -vE '^/nix/store/[^/]*(clang|cctools|xcbuild|compiler-rt|binutils|libcxx)[^/]*/bin' \
    | paste -sd: -)"
  export PATH
fi

if [ "$#" -gt 0 ]; then
  exec "$@"
fi
