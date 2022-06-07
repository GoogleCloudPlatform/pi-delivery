#!/usr/bin/env ruby

# Here's a one-liner to convert the results to csv.
# find . -name 'result-*.txt' -exec sh -c "grep -E '(Far Memory)|(Sequential)|(Threshold)|(Computation)|(Disk I/O)' {} | 
#   grep -Eo '[0-9]+\.[0-9]+ GiB/s' | sed -n '2p;4p;6p;8p;9p;10p' | grep -Eo '[0-9]+\.[0-9]+' | paste -s -d, - " \;

require 'erb'

CONFIG_FILE='y-bench.cfg'
RESULTS_DIR='./bench-results'

def test(count:)
  bytes_per_seek = 256 * 1024 * count
  template = ERB.new(File.read("bench-templ.cfg.erb"))
  cfg_file = "#{RESULTS_DIR}/bench-#{count}.cfg"
  File.write(cfg_file, template.result(binding))
  system("cd y-cruncher && ./y-cruncher config ../#{cfg_file} | tee ../#{RESULTS_DIR}/result-#{count}.txt")
end


(32..72).step(2) do |n|
  test(count: n)
end
