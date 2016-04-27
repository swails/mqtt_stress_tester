#!/usr/bin/env python

import random
try:
    from string import ascii_letters, digits
    chars = ascii_letters + digits
except ImportError:
    from string import letters, digits
    chars = letters + digits

with open('rawtext.passwd', 'w') as f:
    print('stresstest:stressmeout', file=f)
    for i in range(1000):
        user = ''.join(random.choice(chars) for i in range(30))
        passwd = ''.join(random.choice(chars) for i in range(30))
        print('%s:%s' % (user, passwd), file=f)
