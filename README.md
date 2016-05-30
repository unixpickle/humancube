# Overview

This is an experiment. My goal is to feed reconstructions from [cubesolv.es](http://cubesolv.es) to an [RNN](https://en.wikipedia.org/wiki/Recurrent_neural_network) and train it to predict the moves a human might make in a given scenario. If it works out, the computer should be able to solve new scrambles it's never seen before.

# Hypothesis

I do not know what the outcome of this project will be, but here are my predictions:

 * The LSTM will struggle to generalize what it's learned.
   * Especially for free-form things like the X-cross.
 * Training will be difficult since different solvers use different algorithms, different methods, etc.
 * I will mess up the cube-state code at least once, even though I'll use [Gocube](https://github.com/unixpickle/gocube).
