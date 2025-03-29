import React from 'react';

const UserList = ({ currentUser, users, onSelect }) => {
  return (
    <div>
      <h3>Contacts</h3>
      {users
        .filter(u => u.username !== currentUser)
        .map(user => (
          <div
            key={user.username}
            style={{ cursor: 'pointer', margin: '8px 0' }}
            onClick={() => onSelect(user)}
          >
            {user.username}
          </div>
        ))
      }
    </div>
  );
};

export default UserList;
