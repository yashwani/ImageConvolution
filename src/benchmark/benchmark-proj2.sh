#!/bin/bash
#
#SBATCH --mail-user=yashwani@uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj2_benchmark
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/yashwani/src-yashwani/src/benchmark
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=200:00

module load golang/1.16.2

datasets=(small mixture big)

threads=(2 4 6 8 12)

for dataset in "${datasets[@]}"
do
  for thread in "${threads[@]}"
  do
     for iter in {1..5}
     do
       go run proj2/editor $dataset pipeline $thread
     done
  done
done

for dataset in "${datasets[@]}"
do
  for thread in "${threads[@]}"
  do
     for iter in {1..5}
     do
       go run proj2/editor $dataset bsp $thread
     done
  done
done

for dataset in "${datasets[@]}"
do
   for iter in {1..5}
   do
     go run proj2/editor $dataset
   done
done
