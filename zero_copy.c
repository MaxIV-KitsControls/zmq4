extern void go_callback(void *data, void *hint);

void my_free (void *data, void *hint)
{
  go_callback(data, hint);
};
