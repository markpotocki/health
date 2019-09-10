#include <time.h>
#include "cpu_timer.h"

//https://www.tutorialspoint.com/c_standard_library/c_function_clock.htm
int getCPUTime() {
    clock_t start_t, end_t, total_t;
    start_t = clock();
    end_t = clock();

    total_t = (end_t - start_t) / CLOCKS_PER_SEC;

    return total_t;
}