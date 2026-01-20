// TestBed.cpp : This file contains the 'main' function. Program execution begins and ends there.
//

#include <iostream>
#include <thread>
#include <chrono>

int main()
{
    std::cout << "Waiting For something odd ....\n";
    for (int i{}; i < 100; i++) {
        std::cout << ".\n";
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }
    std::cout << "GoodBye\n";
    return 0;
}