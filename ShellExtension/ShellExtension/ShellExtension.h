#pragma once
#include <windows.h>
#include <shlobj.h>
#include <string>

class CShellExt : public IShellExtInit, public IContextMenu {
public:
    CShellExt();
    virtual ~CShellExt();

    // IUnknown
    IFACEMETHODIMP QueryInterface(REFIID riid, void** ppv);
    IFACEMETHODIMP_(ULONG) AddRef();
    IFACEMETHODIMP_(ULONG) Release();

    // IShellExtInit
    IFACEMETHODIMP Initialize(LPCITEMIDLIST, IDataObject*, HKEY);

    // IContextMenu
    IFACEMETHODIMP QueryContextMenu(HMENU, UINT, UINT, UINT, UINT);
    IFACEMETHODIMP InvokeCommand(LPCMINVOKECOMMANDINFO);
    IFACEMETHODIMP GetCommandString(UINT_PTR, UINT, UINT*, LPSTR, UINT);

private:
    long m_cRef;
    std::wstring m_selectedFile;
};
