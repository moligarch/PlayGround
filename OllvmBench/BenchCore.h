#pragma once

#include <string>
#include <Windows.h>


// Metric container
struct BenchResult {
    std::string Name;
    long long DurationMS;
    SIZE_T PeakMemoryKB;
    long long KernelTimeMS;
    long long UserTimeMS;
};

class BenchSuite {
public:
    // Main entry point to run all tests
    static void RunAll(const std::string& buildLabel);

private:
    // Internal helpers
    static BenchResult RunSingleTest(const std::string& name, void(*func)());
    static void PrintResult(const BenchResult& res);

    // Workloads
    static void Workload_MatrixMult();  // Number Cruncher
    static void Workload_Branching();   // Branch Monster
    static void Workload_Memory();      // Memory Thrasher
    static void Workload_IO();          // I/O Blast
    static void Workload_Crypto();      // Crypto Sim
};