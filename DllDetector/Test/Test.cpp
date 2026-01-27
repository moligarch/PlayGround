#include "CppUnitTest.h"
#include <DllDetector/DllDetector.h>
#include "helper.h"

namespace NoAnomalyCheckTests
{
    TEST_CLASS(IsDLL)
    {
    public:
        TEST_METHOD(kernel32_x64_dll)
        {
            auto filePath = GetTestFilePath(L"kernel32_x64.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"kernel32_x64.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(kernel32_x86_dll)
        {
            auto filePath = GetTestFilePath(L"kernel32_x86.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"kernel32_x86.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(tzautoupdate_dll)
        {
            auto filePath = GetTestFilePath(L"tzautoupdate.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"tzautoupdate.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(imaadp32_acm_mui)
        {
            auto filePath = GetTestFilePath(L"imaadp32.acm.mui");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"imaadp32.acm.mui should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(msadp32_acm_mui)
        {
            auto filePath = GetTestFilePath(L"msadp32.acm.mui");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"msadp32.acm.mui should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };

    TEST_CLASS(IsEXE)
    {
    public:
        TEST_METHOD(cmd_x64_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x64.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"cmd_x64.exe should be IS_EXE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(cmd_x64_packed_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x64_packed.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"cmd_x64_packed.exe should be IS_EXE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(cmd_x86_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x86.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"cmd_x86.exe should be IS_EXE.");
            Assert::IsFalse(result.anomalies.empty(), L"As cmd_x86.exe has export table even its' exe, there should be Anomalies here.");
        }

        TEST_METHOD(PackagedCWALauncher_exe)
        {
            auto filePath = GetTestFilePath(L"PackagedCWALauncher.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"PackagedCWALauncher.exe should be IS_EXE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };

    TEST_CLASS(Other)
    {
    public:

        TEST_METHOD(NonExistentFile)
        {
            std::wstring filePath = L"C:\\path\\that\\certainly\\does\\not\\exist\\file.dll";
            auto result = PeAnalysis::IsFileDllW(filePath);
            Assert::AreEqual((int)PeAnalysis::Verdict::FILE_ERROR, (int)result.verdict, L"A non-existent file should result in FILE_ERROR.");
        }

        TEST_METHOD(MalformedDll)
        {
            auto filePath = GetTestFilePath(L"Microsoft.Build.Utilities.Core.ni.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"Microsoft.Build.Utilities.Core.ni.dll should be IS_DLL.");
            Assert::IsFalse(result.anomalies.empty(), L"There should be Header_Contradiction_Dll_Flag_Vs_Entrypoint anomaly");
        }

        TEST_METHOD(MalformedExe)
        {
            auto filePath = GetTestFilePath(L"aitstatic.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"aitstatic.exe should be IS_EXE.");
            Assert::IsFalse(result.anomalies.empty(), L"There should be Header_Contradiction_Exe_with_Export anomaly");
        }

        TEST_METHOD(NotPE_ShortFileBin)
        {
            auto filePath = GetTestFilePath(L"too_short.bin");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"too_short.bin should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(NotPE_EmptyFile)
        {
            auto filePath = GetTestFilePath(L"empty.bin");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"empty.bin should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(NotPE_BatchFile)
        {
            auto filePath = GetTestFilePath(L"tzautoupdate.dat");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring());
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"tzautoupdate.dat should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };
}


namespace EntropyCheckTests
{
    PeAnalysis::Config cfg{
        .options = PeAnalysis::Options::ENTROPY_CHECK,
        .maxFileSizeInMB = 50
    };

    TEST_CLASS(IsDLL)
    {
    public:
        TEST_METHOD(kernel32_x64_dll)
        {
            auto filePath = GetTestFilePath(L"kernel32_x64.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"kernel32_x64.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(kernel32_x86_dll)
        {
            auto filePath = GetTestFilePath(L"kernel32_x86.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"kernel32_x86.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(tzautoupdate_dll)
        {
            auto filePath = GetTestFilePath(L"tzautoupdate.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"tzautoupdate.dll should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(imaadp32_acm_mui)
        {
            auto filePath = GetTestFilePath(L"imaadp32.acm.mui");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"imaadp32.acm.mui should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(msadp32_acm_mui)
        {
            auto filePath = GetTestFilePath(L"msadp32.acm.mui");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_DLL, (int)result.verdict, L"msadp32.acm.mui should be IS_DLL.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };

    TEST_CLASS(IsEXE)
    {
    public:
        TEST_METHOD(cmd_x64_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x64.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"cmd_x64.exe should be IS_EXE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(cmd_x64_packed_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x64_packed.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::MALFORMED, (int)result.verdict, L"cmd_x64_packed.exe should be MALFORMED.");
            Assert::AreEqual(result.anomalies.size(), size_t(3), L"As this sample is packed, there should be 3 anomalies: entry point, packed section, high entropy");
        }

        TEST_METHOD(cmd_x86_exe)
        {
            auto filePath = GetTestFilePath(L"cmd_x86.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::MALFORMED, (int)result.verdict, L"cmd_x86.exe should be MALFORMED.");
            Assert::IsFalse(result.anomalies.empty(), L"As cmd_x86.exe has export table even its' exe, there should be Anomalies here.");
        }

        TEST_METHOD(PackagedCWALauncher_exe)
        {
            auto filePath = GetTestFilePath(L"PackagedCWALauncher.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::IS_EXE, (int)result.verdict, L"PackagedCWALauncher.exe should be IS_EXE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };

    TEST_CLASS(Other)
    {
    public:

        TEST_METHOD(NonExistentFile)
        {
            std::wstring filePath = L"C:\\path\\that\\certainly\\does\\not\\exist\\file.dll";
            auto result = PeAnalysis::IsFileDllW(filePath, cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::FILE_ERROR, (int)result.verdict, L"A non-existent file should result in FILE_ERROR.");
        }

        TEST_METHOD(MalformedDll)
        {
            auto filePath = GetTestFilePath(L"Microsoft.Build.Utilities.Core.ni.dll");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::MALFORMED, (int)result.verdict, L"Microsoft.Build.Utilities.Core.ni.dll should be MALFORMED.");
            Assert::IsFalse(result.anomalies.empty(), L"There should be Header_Contradiction_Dll_Flag_Vs_Entrypoint anomaly");
        }

        TEST_METHOD(MalformedExe)
        {
            auto filePath = GetTestFilePath(L"aitstatic.exe");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::MALFORMED, (int)result.verdict, L"aitstatic.exe should be MALFORMED.");
            Assert::IsFalse(result.anomalies.empty(), L"There should be Header_Contradiction_Exe_with_Export anomaly");
        }

        TEST_METHOD(NotPE_ShortFileBin)
        {
            auto filePath = GetTestFilePath(L"too_short.bin");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"too_short.bin should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(NotPE_EmptyFile)
        {
            auto filePath = GetTestFilePath(L"empty.bin");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"empty.bin should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }

        TEST_METHOD(NotPE_BatchFile)
        {
            auto filePath = GetTestFilePath(L"tzautoupdate.dat");
            AssertTestFileExists(filePath);
            auto result = PeAnalysis::IsFileDllW(filePath.wstring(), cfg);
            Assert::AreEqual((int)PeAnalysis::Verdict::NOT_A_PE, (int)result.verdict, L"tzautoupdate.dat should be NOT_A_PE.");
            Assert::IsTrue(result.anomalies.empty(), L"With no option in config, there should be no anomaly detection.");
        }
    };
}