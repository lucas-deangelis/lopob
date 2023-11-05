# lossless PNG optimizer benchmark

I have a whole lot of PNG files that I keep around for archival, I'd like to reduce their size as much as possible, using a reasonable amount of compute/power/manpower, and without changing the actual image contained in the file.

## Context, what does "losslesss" means here, and other formalities, in the presence of the PNG Specification (Third Edition)

From my understanding of the [4 Concepts](https://www.w3.org/TR/png-3/#4Concepts) chapter of the [PNG Specification](https://www.w3.org/TR/png-3/), I want the reference image to stay the same between the original PNG and the optimized PNG. However a reference image only exists conceptually, so we'll fall back on comparing PNG images.

One issue I already see is that different people may mean different thing by lossless. From the standard ([here](https://www.w3.org/TR/png-3/#4Concepts.Introduction)) (emphasis mine):

> When every pixel is either fully transparent or fully opaque, the alpha separation, alpha compaction, and indexing transformations can cause the recovered reference image to have an alpha sample depth different from the original reference image, or to have no alpha channel. This has no effect on the degree of opacity of any pixel. The two reference images are considered equivalent, and the transformations are considered lossless. *Encoders that nevertheless wish to preserve the alpha sample depth may elect not to perform transformations that would alter the alpha sample depth.*

The definition of the standard: [lossless](https://www.w3.org/TR/png-3/#dfn-lossless)

> method of data compression that permits reconstruction of the original data exactly, bit-for-bit

We now need a way to evaluate that. Idedally multiple ways, if one of them has a bug.

The first is [imagemagick](https://imagemagick.org/), with the `compare` cli tool: if `compare -metric AE image.png optimized-image.png` prints `0`, we'll consider that they are the same.

The second is a very naive Go program that takes two paths as parameters and:

- opens the files
- decode them as PNG images
- checks that their bounds are equal
- go through the two images pixel by pixel and compare them

## Tools

- ect
- oxipng
- optipng
- pngcrush

Measures to do:

- `ect -{1..9} --strict`
- `oxipng -o {0..6}`

## Some terrible, terrible news

In [2. Scope](https://www.w3.org/TR/png-3/#1Scope)

> This specification specifies a datastream and an associated file format, Portable Network Graphics (PNG, pronounced "ping"), for a [lossless](https://www.w3.org/TR/png-3/#dfn-lossless), portable, compressed individual computer graphics image or frame-based animation, transmitted across the Internet.

## Program structure

### v1

On a very basic level, we want/need that takes as an input a list of `(command_to_run, target_file_path)`, run `command_to_run` on `target_file_path` (measuring the execution time), check that the new file is smaller or equal to the original file, check that the two PNG images are the same, and then return `(new_size, execution_time)`. We can start by writing everything to the console.

## Building a shed for my bikes

At first the benchmark was made using python, which kind of workeds well, but at some point I realized that I'd like some concurrency (or at least not the ultra linearity of doing every computation one after the other as it's written in the code), so I'll switch over to Go, because I know Go better.
