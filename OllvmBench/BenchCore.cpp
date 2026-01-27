#include "BenchCore.h"

#include <algorithm>
#include <chrono>
#include <fstream>
#include <iostream>
#include <random>
#include <string>
#include <vector>

#include <windows.h>
#include <psapi.h>

// --- Helper to get process metrics ---
void GetMetrics(HANDLE hProcess, SIZE_T& peakMem, long long& kTime, long long& uTime) {
    PROCESS_MEMORY_COUNTERS pmc;
    if (GetProcessMemoryInfo(hProcess, &pmc, sizeof(pmc))) {
        peakMem = pmc.PeakWorkingSetSize / 1024; // KB
    }

    FILETIME ftCreation, ftExit, ftKernel, ftUser;
    if (GetProcessTimes(hProcess, &ftCreation, &ftExit, &ftKernel, &ftUser)) {
        // FILETIME is in 100-nanosecond intervals. Divide by 10,000 for ms.
        ULARGE_INTEGER k, u;
        k.LowPart = ftKernel.dwLowDateTime; k.HighPart = ftKernel.dwHighDateTime;
        u.LowPart = ftUser.dwLowDateTime; u.HighPart = ftUser.dwHighDateTime;
        kTime = k.QuadPart / 10000;
        uTime = u.QuadPart / 10000;
    }
}

void BenchSuite::RunAll(const std::string& buildLabel) {
    std::cout << "=============================================================\n";
    std::cout << " BENCHMARK SUITE: " << buildLabel << "\n";
    std::cout << "=============================================================\n";
    std::cout << "Test Name           | Time(ms) | Kernel(ms) | User(ms) | RAM(KB)\n";
    std::cout << "-------------------------------------------------------------\n";

    // Run workloads sequentially
    PrintResult(RunSingleTest("MatrixMult (INT)", Workload_MatrixMult));
    PrintResult(RunSingleTest("Branching (Rec)", Workload_Branching));
    PrintResult(RunSingleTest("Memory (Rand)", Workload_Memory));
    PrintResult(RunSingleTest("Crypto (Sim)", Workload_Crypto));
    PrintResult(RunSingleTest("File I/O", Workload_IO));

    std::cout << "=============================================================\n\n";
}

BenchResult BenchSuite::RunSingleTest(const std::string& name, void(*func)()) {
    HANDLE hSelf = GetCurrentProcess();

    // Baseline metrics before run
    SIZE_T memStart = 0, memEnd = 0;
    long long kStart = 0, kEnd = 0;
    long long uStart = 0, uEnd = 0;

    GetMetrics(hSelf, memStart, kStart, uStart);

    auto start = std::chrono::high_resolution_clock::now();

    // EXECUTE WORKLOAD
    func();

    auto end = std::chrono::high_resolution_clock::now();

    GetMetrics(hSelf, memEnd, kEnd, uEnd);

    BenchResult res;
    res.Name = name;
    res.DurationMS = std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();
    res.PeakMemoryKB = memEnd; // Process memory usually grows, so current peak is valid
    res.KernelTimeMS = kEnd - kStart;
    res.UserTimeMS = uEnd - uStart;

    return res;
}

void BenchSuite::PrintResult(const BenchResult& r) {
    // Basic formatting for alignment
    printf("%-19s | %8lld | %10lld | %8lld | %7zu\n",
        r.Name.c_str(), r.DurationMS, r.KernelTimeMS, r.UserTimeMS, r.PeakMemoryKB);
}

// ----------------------------------------------------------------
// 1. Matrix Multiplication (Integer Math heavy)
// ----------------------------------------------------------------
void BenchSuite::Workload_MatrixMult() {
    const int N = 500; // 500x500 matrix
    std::vector<int> A(N * N, 1), B(N * N, 2), C(N * N, 0);

    for (int i = 0; i < N; ++i) {
        for (int j = 0; j < N; ++j) {
            for (int k = 0; k < N; ++k) {
                C[i * N + j] += A[i * N + k] * B[k * N + j];
            }
        }
    }
}

// ----------------------------------------------------------------
// 2. Branching (Recursion + Logic)
// ----------------------------------------------------------------
long long Fibonacci(int n) {
    if (n <= 1) return n;
    return Fibonacci(n - 1) + Fibonacci(n - 2);
}

void BenchSuite::Workload_Branching() {
    volatile long long result = 0;
    // Calculate 35th fib number multiple times
    for (int i = 0; i < 5; i++) {
        result += Fibonacci(35);
    }
}

// ----------------------------------------------------------------
// 3. Memory Thrasher (Random Access)
// ----------------------------------------------------------------
void BenchSuite::Workload_Memory() {
    const size_t SIZE = 10 * 1024 * 1024; // 10 Million ints (~40MB)
    std::vector<int> data(SIZE);

    // Linear fill
    for (size_t i = 0; i < SIZE; ++i) data[i] = (int)i;

    // Random shuffle (heavy memory movement)
    std::random_device rd;
    std::mt19937 g(rd());
    std::shuffle(data.begin(), data.end(), g);

    // Random Access Read
    volatile long long sum = 0;
    for (size_t i = 0; i < SIZE; ++i) {
        sum += data[data[i] % SIZE]; // Chasing pointers essentially
    }
}

// ----------------------------------------------------------------
// 4. Crypto Simulation (Bitwise ops)
// ----------------------------------------------------------------
void BenchSuite::Workload_Crypto() {
    const size_t BUF_SIZE = 1024 * 1024 * 5; // 5MB buffer
    std::vector<unsigned char> buffer(BUF_SIZE, 0xAB);
    unsigned char key = 0x5C;

    // Multiple rounds of XOR/Rotate
    for (int round = 0; round < 50; ++round) {
        for (size_t i = 0; i < BUF_SIZE; ++i) {
            buffer[i] ^= key;
            buffer[i] = (buffer[i] << 1) | (buffer[i] >> 7); // Rotate Left
            key = (key + i) & 0xFF; // Mutating key
        }
    }
}

// ----------------------------------------------------------------
// 5. I/O Blast
// ----------------------------------------------------------------
void BenchSuite::Workload_IO() {
    const std::string filename = "bench_temp.dat";
    const int LINES = 50000;

    // Write
    {
        std::ofstream out(filename);
        for (int i = 0; i < LINES; ++i) {
            out << "Line " << i << ": This is some junk data to test IO throughput.\n";
        }
    } // Flush and close

    // Read back
    {
        std::ifstream in(filename);
        std::string line;
        volatile size_t count = 0;
        while (std::getline(in, line)) {
            count += line.length();
        }
    }

    // Cleanup
    remove(filename.c_str());
}