#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <unistd.h>

void* fast_function()
{
  while (1) {
    printf("Fast function \n");
    usleep(100);
  }
}

void* slow_function()
{
  while (1) {
    printf("Slow function \n");
    usleep(400);
  }
}

int main()
{
  pthread_t thread1;
  pthread_t thread2;
  int thread_rc;

  if ((thread_rc=pthread_create(&thread1,NULL,fast_function,NULL))!=0)
  {
    printf("Error creating the thread. Code %i",thread_rc);
    return -1;
  }
  //  int *ptr_output_data;


  if ((thread_rc=pthread_create(&thread2,NULL,slow_function,NULL))!=0)
  {
    printf("Error creating the thread. Code %i",thread_rc);
    return -1;
  }
  //  int *ptr_output_data2;
   pthread_join(thread1,NULL);
  pthread_join(thread2,NULL);

  return 0;
}
