#!/usr/bin/env bash

BASE_DIR="$(pwd)"
ASSETS=$BASE_DIR/assets
EXAMPLE_ASSETS=$ASSETS/examples

echo "Remove existing example outputs"
rm -f $EXAMPLE_ASSETS/*

echo "Generate example outputs"

goimportcycle -path $BASE_DIR/examples/importcycle/ -resolution file -dot $EXAMPLE_ASSETS/file.dot
dot -Tpng -o $EXAMPLE_ASSETS/file.png $EXAMPLE_ASSETS/file.dot

goimportcycle -path $BASE_DIR/examples/importcycle/ -resolution package -dot $EXAMPLE_ASSETS/package.dot
dot -Tpng -o $EXAMPLE_ASSETS/package.png $EXAMPLE_ASSETS/package.dot
