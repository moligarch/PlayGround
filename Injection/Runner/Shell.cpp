#include "Shell.h"
#include <Windows.h>

#include <iostream>

static void __stdcall __StartInjFunc() {
    //std::cout << "If You are seeing this, injection succeed" << std::endl;
    MessageBox(NULL, L"Test", L"Test", 0);
}
static __declspec() void __EndInjFunc() {
    return;
}




bool InjectIntoProcess(DWORD pid)
{
    HANDLE hProcess = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_CREATE_THREAD, FALSE, pid);
    DWORD dwFuncSize = ((DWORD)__EndInjFunc - (DWORD)__StartInjFunc);
    if (dwFuncSize != 0 && hProcess != NULL)
    {
        DWORD AllocSize = dwFuncSize + sizeof(DWORD);
        DWORD  dwOldAttrib = NULL;
        LPVOID vpRemoteMemory = VirtualAllocEx(hProcess, NULL, AllocSize, MEM_RESERVE | MEM_COMMIT, PAGE_READWRITE);
        if (vpRemoteMemory != NULL)
        {
            if (WriteProcessMemory(hProcess, vpRemoteMemory, (LPCVOID)__StartInjFunc, dwFuncSize, NULL))
            {
                if (VirtualProtectEx(hProcess, vpRemoteMemory, AllocSize, PAGE_EXECUTE_READWRITE, &dwOldAttrib)) {
                    DWORD dwThreadId = NULL;
                    if (CreateRemoteThread(hProcess, NULL, NULL, (LPTHREAD_START_ROUTINE)vpRemoteMemory, NULL, NULL, &dwThreadId) != NULL) {
                        return true;
                    }
                }
            }
        }
    }
    std::cout << GetLastError() << std::endl;
    return false;
}
