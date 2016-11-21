CONFIG = {}

def read_config
  while true
    line = $stdin.readline.strip
    break if line == "compl|"
    comp = line.split("|")
    next unless comp[0] == "cnf"
    CONFIG[comp[1]] = comp[2..-1].join
  end
end


read_config

x = Thread.new do
  while true do
    read_config
    puts "cnf|time|#{CONFIG["time"].to_i * -1}"
    STDOUT.flush
  end
end

x.join
