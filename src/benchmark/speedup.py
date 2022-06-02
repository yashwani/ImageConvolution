import matplotlib
import sys
import collections

matplotlib.use("agg")
import matplotlib.pyplot

path = sys.argv[1]
# filepath = './slurm/out/51410.slurm3.stdout'
filepath = './slurm/out/' + path

file = open(filepath, 'r').read().splitlines()


dataset = ["small", "mixture","big"]
threads = [2,4,6,8,12]
iter = 5

## average each set of 5
avg_file = []

tot = 0
for i in range(len(file)):
    tot += float(file[i])
    if (i+1)%5 == 0:
        avg_file.append(tot/5)
        tot = 0

pipeline_graph = collections.defaultdict(list)
bsp_graph = collections.defaultdict(list)
sequential = {}

for i in range(0,5):
    pipeline_graph["small"].append(avg_file[i])
for i in range(5,10):
    pipeline_graph["mixture"].append(avg_file[i])
for i in range(10,15):
    pipeline_graph["big"].append(avg_file[i])

for i in range(15,20):
    bsp_graph["small"].append(avg_file[i])
for i in range(20,25):
    bsp_graph["mixture"].append(avg_file[i])
for i in range(25,30):
    bsp_graph["big"].append(avg_file[i])


sequential["small"] = avg_file[30]
sequential["mixture"] = avg_file[31]
sequential["big"] = avg_file[32]

x_vals = threads

#################### Pipeline

matplotlib.pyplot.figure()
for label, y_vals in pipeline_graph.items():
    y_vals = list(map(lambda x: sequential[label]/x, y_vals))
    matplotlib.pyplot.plot(x_vals, y_vals, label = label)


matplotlib.pyplot.legend()

# naming the x axis
matplotlib.pyplot.xlabel('Number of Threads')
# naming the y axis
matplotlib.pyplot.ylabel('Speedup')

# giving a title to my graph
matplotlib.pyplot.title('Speedup Graph Pipeline')

# function to show the plot
matplotlib.pyplot.savefig('speedup-pipeline')


#################### BSP

matplotlib.pyplot.figure()
for label, y_vals in bsp_graph.items():
    y_vals = list(map(lambda x: sequential[label]/x, y_vals))
    matplotlib.pyplot.plot(x_vals, y_vals, label = label)


matplotlib.pyplot.legend()

# naming the x axis
matplotlib.pyplot.xlabel('Number of Threads')
# naming the y axis
matplotlib.pyplot.ylabel('Speedup')

# giving a title to my graph
matplotlib.pyplot.title('Speedup Graph BSP')

# function to show the plot
matplotlib.pyplot.savefig('speedup-bsp')