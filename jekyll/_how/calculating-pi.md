---
title: Calculating Pi
position_number: 2
---

In 2019, [Emma Haruka Iwao](https://twitter.com/Yuryu) broke the Pi digits world record by computing 31,415,926,535,897 digits of Pi using Google Cloud Platform! Learn more details from Emma's [Calculating a record-breaking 31.4 trillion digits](https://cloud.google.com/blog/products/compute/calculating-31-4-trillion-digits-of-archimedes-constant-on-google-cloud) blog, and Alex Yee (creator of y-cruncher)'s [Google Cloud Topples the Pi record](http://www.numberworld.org/blogs/2019_3_14_pi_record/) article.

Back in 2017, we used [y-cruncher](http://www.numberworld.org/y-cruncher/) to calculate 750 billion digits of Pi on Google Compute Engine. We used a 64-core instance, with 416GB of RAM, and tons of Local SSD + Persistent SSD.

<blockquote class="twitter-tweet" data-lang="en"><p lang="en" dir="ltr">Sometimes, you just need a big honkin&#39; VM <a href="https://twitter.com/googlecloud">@googlecloud</a> <a href="https://t.co/WUtLIhW4bs">pic.twitter.com/WUtLIhW4bs</a></p>&mdash; Greg Wilson (@gregsramblings) <a href="https://twitter.com/gregsramblings/status/838485706230087680">March 5, 2017</a></blockquote>
<script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

At its peak CPU utilization, it consumed all 64 cores, and 5+TB of SSD storage.

<blockquote class="twitter-tweet" data-lang="en"><p lang="en" dir="ltr">Consumed all 64 cores on a <a href="https://twitter.com/googlecloud">@googlecloud</a> vm, of 8x375GB Local SSD &amp; using 382GB of 416GB RAM for some serious computation <a href="https://t.co/Fzptt1nDCw">pic.twitter.com/Fzptt1nDCw</a></p>&mdash; Ray Tsang (@saturnism) <a href="https://twitter.com/saturnism/status/838791108016533505">March 6, 2017</a></blockquote>
<script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

In 2018, we were able to recalculate 750 billion digits in just 10 hours!

<blockquote class="twitter-tweet" data-lang="en"><p lang="en" dir="ltr">Happy <a href="https://twitter.com/hashtag/PiDay?src=hash&amp;ref_src=twsrc%5Etfw">#PiDay</a>! We just calculated 750-billion digits of Pi calculation in 10 hours (compared to 2-days it took in 2017) using something we&#39;ll announce soon on <a href="https://twitter.com/GCPcloud?ref_src=twsrc%5Etfw">@GCPcloud</a>. Stayed tuned!  <a href="https://twitter.com/hashtag/PiDay2018?src=hash&amp;ref_src=twsrc%5Etfw">#PiDay2018</a> <a href="https://t.co/lFfSv63EzL">pic.twitter.com/lFfSv63EzL</a></p>&mdash; Ray Tsang (@saturnism) <a href="https://twitter.com/saturnism/status/973954547872862209?ref_src=twsrc%5Etfw">March 14, 2018</a></blockquote>
<script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

We also calculated 1 trillion digits of Pi using a 96-core 14TB `n1-megamem-96` machine.

<blockquote class="twitter-tweet" data-lang="en"><p lang="en" dir="ltr">Near full utilization of a 96-cores &amp; 1.4TB of RAM n1-megamem-96 instance on <a href="https://twitter.com/GCPcloud?ref_src=twsrc%5Etfw">@GCPcloud</a> <a href="https://t.co/C6JxMHrrzj">pic.twitter.com/C6JxMHrrzj</a></p>&mdash; Ray Tsang (@saturnism) <a href="https://twitter.com/saturnism/status/973619880024043521?ref_src=twsrc%5Etfw">March 13, 2018</a></blockquote>
<script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

