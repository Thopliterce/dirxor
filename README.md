dirxor
---

A proof of concept to demonstrate a loophole in existing copyright systems.

# Usage
```plain
dirxor [-i DIR | -o DIR]...
-i DIR   Add an input directory or file
-o DIR   Add an output directory or file
```

Example:
```bash
# Assume that you have a directory called "dir1".

# Creates two directories called "dir2" and "dir3" with identical structure.
# The files will have exact same sizes, and they'll be filled with data that
# is mathematically indistinguishable from pure nonsense on its own.
# You may specify more than two output directories here, but all of them
# need to be supplied as inputs to the next command in order to recover
# the original directory.
dirxor -i dir1 -o dir2 -o dir3

# However, when both directories are merged together, it results in the
# orignal data.
# You should specify only one output here, otherwise you'll get garbage.
dirxor -i dir2 -i dir3 -o dir4
# Now dir4 will contain the same contents as dir1.
```

# How does it work?

Original thread: https://www.reddit.com/r/Piracy/comments/dy5y4j/piracy\_that\_cant\_even\_be\_sued/

Assume you want to distribute a pirated file called A. You create a random file called B, xor it with the original file A, and produce A^B. A^B and B are both completely random and unrelated to A on their own, so you can safely distribute both files since it is mathematically impossible to relate any one of the two files to A in any way. However, when they are XOR'ed together (i.e. computing (A^B)^B) you get A, which is the original file.

This means that you can distribute the file B and A^B on two different sites, and since both files are random content, neither can be sued for copyright violation. However, anyone who downloads both files will be able to recover the original file A.

This is different from encrypting pirated content with a password, since with the password anyone can prove that the content is indeed copyrighted. In this case, the fact that xor'ing B and A^B produces copyrighted content A is purely coincidental; even knowing the presence of such a correlation will not prove that either of the two files contains copyrighted content (otherwise for any random data B you can distribute A^B and then claim B is copyrighted since B and A^B are both random and mathematically indistinguishable). The drawback is that 2x more traffic is needed to download the original file. The above procedure is similar to the one-time-pad scheme, but any information-theoretically-secure methods can be used to distribute pirated files in a similar fashion.
