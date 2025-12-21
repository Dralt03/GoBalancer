import subprocess
import re
import matplotlib.pyplot as plt
import sys
import os
import json

def run_benchmarks():
    print("Running benchmarks...")
    # Run benchmarks specifically for the Pick method in balancer package
    cmd = ["go", "test", "-bench=Pick", "./internal/balancer/...", "-v", "-benchmem"]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print("Error running benchmarks:")
        print(result.stderr)
        sys.exit(1)
    return result.stdout

def parse_results(output):
    # Regex to match: Benchmark<Algo>_Pick/<Count>-<Cores> <Iterations> <NsPerOp> ns/op
    # Example: BenchmarkRoundRobin_Pick/1000-16         1000000              2500 ns/op
    pattern = re.compile(r"Benchmark(\w+)_Pick/(\d+)-\d+\s+\d+\s+(\d+(\.\d+)?)\s+ns/op")
    
    data = {}
    
    for line in output.splitlines():
        match = pattern.search(line)
        if match:
            algo = match.group(1)
            count = int(match.group(2))
            ns_op = float(match.group(3))
            
            if algo not in data:
                data[algo] = []
            
            data[algo].append((count, ns_op))
            
    for algo in data:
        data[algo].sort(key=lambda x: x[0])
        
    return data

def save_results_to_json(data):
    # Convert tuples to lists for JSON compatibility if needed, 
    # but (count, ns_op) tuples usually verify fine.
    # We might want to clear up the structure for easier reading.
    output_file = "benchmark_data.json"
    with open(output_file, "w") as f:
        json.dump(data, f, indent=4)
    print(f"Results saved to {output_file}")

def plot_results(data):
    print("Plotting results...")
    plt.figure(figsize=(12, 8))
    
    for algo, points in data.items():
        counts = [p[0] for p in points]
        times = [p[1] for p in points]
        plt.plot(counts, times, marker='o', label=algo)
        
    plt.xscale('log')
    plt.xlabel('Number of Backends (log scale)')
    plt.ylabel('Time per Operation (ns)')
    plt.title('Load Balancer Algorithm Performance vs Backend Count')
    plt.legend()
    
    output_file = "benchmark_results.png"
    plt.savefig(output_file)
    print(f"Graph saved to {output_file}")

def main():
    # Ensure we are in the project root
    if not os.path.exists("go.mod"):
        print("Error: Please run this script from the project root directory")
        sys.exit(1)
        
    output = run_benchmarks()
    data = parse_results(output)
    
    if not data:
        print("No benchmark data found.")
        sys.exit(1)
        
    save_results_to_json(data)
    plot_results(data)

if __name__ == "__main__":
    main()
