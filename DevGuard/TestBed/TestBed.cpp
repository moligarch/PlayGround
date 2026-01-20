// TestBed.cpp : This file contains the 'main' function. Program execution begins and ends there.
//

#include <iostream>
#include "DevGuard/strings.h"


int main()
{
#if defined(_MSVC_LANG) && (_MSVC_LANG >= 202002L) // C++ >= 20
    auto a = "ascii Hello World!"_sec;
    auto w = L"wide Hello World!"_sec;
    auto u16 = u"utf16 Hello World!"_sec;
    auto u32 = U"utf32 Hello World!"_sec;
    std::cout << a << std::endl;
    std::wcout << w << std::endl;
#else
    auto macro = DVGRD_PROTECT_STR("macro Hello World!");
    auto macroW = DVGRD_PROTECT_STR(L"macro wide Hello World!");
    std::cout << macro << std::endl;
    std::wcout << macroW << std::endl;
#endif

    auto test = "plain Hello World";
    std::wcout << test << std::endl;
    return 0;
}