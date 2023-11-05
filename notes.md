# lossless PNG optimizer benchmark

## Context, what does "losslesss" means here, and other formalities, in the presence of the PNG Specification (Third Edition)

I have a whole lot of PNG files that I keep around for archival, I'd like to reduce their size as much as possible, using a reasonable amount of compute/power, and without changing the actual image contained in the file. From my understanding of the [4 Concepts](https://www.w3.org/TR/png-3/#4Concepts) chapter of the [PNG Specification](https://www.w3.org/TR/png-3/), I want at least the reference image to stay the same between the original PNG and the optimized PNG.

One issue I already see is that different people may mean different thing by lossless. From the standard ([here](https://www.w3.org/TR/png-3/#4Concepts.Introduction)) (emphasis mine):

> When every pixel is either fully transparent or fully opaque, the alpha separation, alpha compaction, and indexing transformations can cause the recovered reference image to have an alpha sample depth different from the original reference image, or to have no alpha channel. This has no effect on the degree of opacity of any pixel. The two reference images are considered equivalent, and the transformations are considered lossless. *Encoders that nevertheless wish to preserve the alpha sample depth may elect not to perform transformations that would alter the alpha sample depth.*

We now need a way to evaluate that. Idedally multiple ways, if one of them has a bug. The first is [imagemagick](https://imagemagick.org/), with the `compare` cli tool: if `compare -metric AE image.png optimized-image.png` prints `0`, we'll consider that they are the same.

## Tools

- ect
- oxipng
- optipng
- pngcrush

Measures to do:

- `ect -{1..9} --strict`
- `oxipng -o {1..6}`