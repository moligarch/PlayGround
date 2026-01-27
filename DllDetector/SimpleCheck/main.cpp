#include <cassert>
#include <filesystem>
#include <iostream>
#include <map>
#include <string>
#include <system_error>

#include "SimplePeParser.h"


/// Verdict for each file
enum class Verdict { DLL, NotDll };

/// Process a single file: updates counters and prints result
void ProcessFile(const fs::path& path, std::map<Verdict, int>& counters) noexcept 
{
    auto strPath = path.string();
    bool isDll = PEParser::IsDll(strPath);
    Verdict v = isDll ? Verdict::DLL : Verdict::NotDll;
    counters[v]++;
    std::cout << strPath << " -> "
        << (isDll ? "DLL" : "Not DLL") << '\n';
}
//////////////////////////////////////////////////////////////


int wmain(int argc, wchar_t* argv[]) {
    if (argc != 2) {
        std::wcerr << "Usage: " << argv[0] << " <path>\n";
        return 1;
    }

    fs::path target{ argv[1] };
    fs::directory_options options = fs::directory_options::skip_permission_denied;
    std::error_code ec;

    if (!fs::exists(target, ec)) {
        std::cerr << "Error: Path does not exist: " << target << '\n';
        return 1;
    }

    std::map<Verdict, int> counters{ {Verdict::DLL,0},{Verdict::NotDll,0} };
    int filesScanned = 0;

    auto handleFile = [&](const fs::path& p) {
        if (fs::is_regular_file(p, ec)) {
            ProcessFile(p, counters);
            ++filesScanned;
        }
        };

    if (fs::is_directory(target, ec)) {
        std::cout << "Scanning directory recursively: " << target << "\n";
        try {
            for (auto const& entry : fs::recursive_directory_iterator(target, options, ec)) {
                handleFile(entry.path());
            }
        }
        catch (const std::exception& e) {
            std::cerr << "Error during traversal: " << e.what() << '\n';
        }
    }
    else {
        handleFile(target);
    }

    std::cout << "\nScanned " << filesScanned << " file(s). \n"
        << "DLLs: " << counters[Verdict::DLL] << ", "
        << "Not DLLs: " << counters[Verdict::NotDll] << "\n";
    return 0;
}

//int main() {
//
//    struct TestCase { const char* path; bool expect; };
//    const TestCase tests[] = {
//        { "../Test/Samples/aitstatic.exe", false },
//        { "../Test/Samples/cmd_x64.exe", false },
//        { "../Test/Samples/cmd_x64_packed.exe", false },
//        { "../Test/Samples/cmd_x86.exe", false },
//        { "../Test/Samples/empty.bin", false },
//        { "../Test/Samples/filelist.txt", false },
//        { "../Test/Samples/imaadp32.acm.mui", true },
//        { "../Test/Samples/kernel32_x64.dll", true },
//        { "../Test/Samples/kernel32_x86.dll", true },
//        { "../Test/Samples/Microsoft.Build.Utilities.Core.ni.dll", true },
//        { "../Test/Samples/msadp32.acm.mui", true },
//        { "../Test/Samples/not_a_pe.txt", false },
//        { "../Test/Samples/PackagedCWALauncher.exe", false },
//        { "../Test/Samples/too_short.bin", false },
//        { "../Test/Samples/tzautoupdate.dat", false },
//        { "../Test/Samples/tzautoupdate.dll", true }
//    };
//
//    std::cout << std::filesystem::current_path().string() << std::endl;
//
//    for (const auto& tc : tests) {
//        bool result = PEParser::IsDll(tc.path);
//        std::cout << tc.path << ": " << (result ? "DLL" : "Not DLL") << "\n";
//        assert(result == tc.expect);
//    }
//
//    std::cout << "All tests passed.\n";
//    return 0;
//}
