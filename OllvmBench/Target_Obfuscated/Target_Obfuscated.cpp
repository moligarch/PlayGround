#include "../BenchCore.h"
#include <iostream>

int main() {
    std::cout << "Starting OBFUSCATED Benchmark...\n\n";

    // Since this project is compiled with OLLVM, the code inside BenchCore
    // (which is compiled as part of this project) will be obfuscated.
    BenchSuite::RunAll("Target: OBFUSCATED (OLLVM -fla -sub -bcf)");

    std::cout << "Press ENTER to exit...";
    std::cin.get();
    return 0;
}